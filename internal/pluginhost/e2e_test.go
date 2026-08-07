package pluginhost_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

const (
	e2eDeployType = "webhook-deployer"
	e2eAccessType = "webhook"
)

var (
	pilotOnce sync.Once
	pilotBin  string
	pilotErr  error
)

func buildPilot(t *testing.T) string {
	t.Helper()
	pilotOnce.Do(func() {
		out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
		if err != nil {
			pilotErr = err
			return
		}
		certimateRoot := strings.TrimSpace(string(out))
		pluginsDir := filepath.Join(certimateRoot, "..", "plugins")
		if _, err := os.Stat(filepath.Join(pluginsDir, "go.mod")); err != nil {
			pilotErr = errors.New("plugins module not found at " + pluginsDir)
			return
		}
		bin := filepath.Join(os.TempDir(), "certimate-e2e-webhook-deployer")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		build := exec.Command("go", "build", "-o", bin, "./webhook-deployer")
		build.Dir = pluginsDir
		if buildOut, err := build.CombinedOutput(); err != nil {
			pilotErr = errors.New(strings.TrimSpace(string(buildOut)) + " | " + err.Error())
			return
		}
		pilotBin = bin
	})
	if pilotErr != nil {
		t.Skipf("pilot build unavailable: %v", pilotErr)
	}
	return pilotBin
}

func setupE2EPluginDir(t *testing.T) string {
	t.Helper()
	bin := buildPilot(t)
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, e2eDeployType)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"version":              "0.1.0",
		"provider_type":        e2eDeployType,
		"access_provider_type": e2eAccessType,
		"display_name_key":     "plugin.webhook-deployer.name",
		"deploy_category":      "other",
		"protocol_version":     plugin.ProtocolVersion,
		"binary":               "webhook-deployer",
	}
	mb, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), mb, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bin, filepath.Join(pluginDir, "webhook-deployer")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestE2E_PilotDeploysThroughFullStack(t *testing.T) {
	dir := setupE2EPluginDir(t)

	var seen struct {
		method string
		path   string
		auth   string
		body   string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.method = r.Method
		seen.path = r.URL.Path
		seen.auth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		seen.body = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	catalog, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir:   dir,
		CoreVersion: "0.4.28",
	}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan/register errors: %+v", errs)
	}

	entries := catalog.Entries()
	if len(entries) != 1 || entries[0].ProviderType != e2eDeployType {
		t.Fatalf("catalog missing pilot entry: %+v", entries)
	}
	if entries[0].Source != pluginhost.SourcePlugin {
		t.Fatalf("entry source = %q", entries[0].Source)
	}

	factory, err := deployers.Registries.Get(domain.DeploymentProviderType(e2eDeployType))
	if err != nil {
		t.Fatalf("pilot not in deployer registry: %v", err)
	}
	access, _ := json.Marshal(map[string]string{"url": srv.URL, "method": "POST", "headers": "Authorization: Bearer e2e-secret"})
	extended, _ := json.Marshal(map[string]string{})

	deployer, err := factory(&deployers.ProviderFactoryOptions{
		ProviderAccessConfig:   jsonMap(access),
		ProviderExtendedConfig: jsonMap(extended),
	})
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	sink := &captureSink{}
	deployer.SetLogger(slog.New(sink))
	certPEM, keyPEM := e2eSelfSignedCert(t, "e2e.example.com")
	if _, err := deployer.Deploy(context.Background(), certPEM, keyPEM); err != nil {
		t.Fatalf("deploy through full stack: %v", err)
	}

	if !sink.has("webhook responded") {
		t.Fatalf("plugin logs did not reach the deployer sink (parity with built-in); records=%+v", sink.records)
	}

	if seen.method != "POST" {
		t.Fatalf("webhook target saw %s", seen.method)
	}
	if seen.auth != "Bearer e2e-secret" {
		t.Fatalf("auth = %q", seen.auth)
	}
	var gotBody map[string]string
	if err := json.Unmarshal([]byte(seen.body), &gotBody); err != nil {
		t.Fatalf("body not json: %v (%s)", err, seen.body)
	}
	if gotBody["cert"] != certPEM {
		t.Fatalf("cert not forwarded: %s", seen.body)
	}
}

func e2eSelfSignedCert(t *testing.T, commonName string) (certPEM, keyPEM string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: commonName},
		DNSNames:     []string{commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM
}

func TestE2E_SchemaEndpointServesPluginEnvelope(t *testing.T) {
	dir := setupE2EPluginDir(t)
	if _, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir: dir, CoreVersion: "0.4.28",
	}, slog.Default()); len(errs) != 0 {
		t.Fatalf("scan/register errors: %+v", errs)
	}

	repo := repository.NewProviderSchemaRepository()
	svc := providerschema.NewProviderSchemaService(repo)

	env, err := svc.GetByProviderType(context.Background(), e2eDeployType)
	if err != nil {
		t.Fatalf("schema endpoint for plugin deploy type: %v", err)
	}
	if env.Provider != e2eDeployType || env.SchemaVersion != providerschema.SchemaVersion {
		t.Fatalf("unexpected envelope: %+v", env)
	}
	if len(env.Schema.Columns) == 0 {
		t.Fatal("deploy envelope has no columns")
	}

	if env, ok := providerschema.Envelopes.GetEnvelope(e2eAccessType); ok || env != nil {
		t.Fatalf("plugin must NOT register an access envelope for the reused builtin %q", e2eAccessType)
	}
}

func TestE2E_IncompatibleVersionRejected(t *testing.T) {
	bin := buildPilot(t)
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "bad-version")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"provider_type":        "bad-version",
		"access_provider_type": "bad-version-access",
		"protocol_version":     999,
		"binary":               "webhook-deployer",
	}
	mb, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), mb, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bin, filepath.Join(pluginDir, "webhook-deployer")); err != nil {
		t.Fatal(err)
	}

	_, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir: dir, CoreVersion: "0.4.28",
	}, slog.Default())
	if len(errs) == 0 {
		t.Fatal("expected incompatible-version plugin to be rejected")
	}
	var inc *plugin.ErrPluginIncompatible
	found := false
	for _, e := range errs {
		if errors.As(e, &inc) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ErrPluginIncompatible, got %+v", errs)
	}
}

func TestE2E_BuiltinStillRegistered(t *testing.T) {
	if _, err := deployers.Registries.Get(domain.DeploymentProviderTypeLocal); err != nil {
		t.Fatalf("builtin LOCAL deployer missing: %v", err)
	}
}

func jsonMap(raw []byte) map[string]any {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		panic(err)
	}
	return m
}

type captureSink struct {
	mu      sync.Mutex
	records []sinkRecord
}

type sinkRecord struct {
	Level   slog.Level
	Message string
}

func (c *captureSink) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (c *captureSink) Handle(_ context.Context, r slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, sinkRecord{Level: r.Level, Message: r.Message})
	return nil
}

func (c *captureSink) WithAttrs(_ []slog.Attr) slog.Handler { return c }
func (c *captureSink) WithGroup(_ string) slog.Handler      { return c }

func (c *captureSink) has(msg string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, r := range c.records {
		if r.Message == msg {
			return true
		}
	}
	return false
}
