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
	"reflect"
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
	pilotTemp string
)

func TestMain(m *testing.M) {
	code := m.Run()
	if pilotTemp != "" {
		_ = os.RemoveAll(pilotTemp)
	}
	os.Exit(code)
}

func buildPilot(t *testing.T) string {
	t.Helper()
	pilotOnce.Do(func() {
		pilotTemp, pilotErr = os.MkdirTemp("", "certimate-e2e-plugin-")
		if pilotErr != nil {
			return
		}
		bin := filepath.Join(pilotTemp, "webhook-deployer")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		build := exec.Command("go", "build", "-o", bin, "github.com/certimate-go/certimate/internal/pluginhost/testplugin")
		if buildOut, err := build.CombinedOutput(); err != nil {
			pilotErr = errors.New(strings.TrimSpace(string(buildOut)) + " | " + err.Error())
			return
		}
		pilotBin = bin
	})
	if pilotErr != nil {
		t.Fatalf("E2E fixture build failed: %v", pilotErr)
	}
	return pilotBin
}

func writeFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupE2EPlugin(t *testing.T, root, providerType, accessType string, fixture map[string]any) string {
	t.Helper()
	dir := filepath.Join(root, providerType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{
		Version: "0.1.0", ProviderType: providerType, AccessProviderType: accessType,
		DisplayNameKey: "plugin.e2e.name", DeployCategory: "other",
		ProtocolVersion: plugin.ProtocolVersion, Binary: "webhook-deployer",
	}
	writeFixtureJSON(t, filepath.Join(dir, "manifest.json"), manifest)
	if fixture == nil {
		fixture = map[string]any{}
	}
	if _, ok := fixture["providerType"]; !ok {
		fixture["providerType"] = providerType
	}
	if _, ok := fixture["accessProviderType"]; !ok {
		fixture["accessProviderType"] = accessType
	}
	if _, ok := fixture["schemaRevision"]; !ok {
		fixture["schemaRevision"] = "v1"
	}
	writeFixtureJSON(t, filepath.Join(dir, "fixture.json"), fixture)
	binary, err := os.ReadFile(buildPilot(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifest.Binary), binary, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func setupE2EPluginDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	setupE2EPlugin(t, dir, e2eDeployType, e2eAccessType, map[string]any{"omitAccessSchema": true})
	return dir
}

func TestE2E_PilotDeploysThroughFullStack(t *testing.T) {
	dir := setupE2EPluginDir(t)

	type observation struct {
		method string
		path   string
		auth   string
		body   string
	}
	received := make(chan observation, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received <- observation{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization"), body: string(b)}
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

	t.Cleanup(catalog.Close)
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

	var seen observation
	select {
	case seen = <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("webhook did not receive deployment")
	}
	if seen.method != "POST" {
		t.Fatalf("webhook target saw %s", seen.method)
	}
	if seen.auth != "Bearer e2e-secret" {
		t.Fatalf("auth = %q", seen.auth)
	}
	var gotBody struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	}
	if err := json.Unmarshal([]byte(seen.body), &gotBody); err != nil {
		t.Fatalf("body not json: %v (%s)", err, seen.body)
	}
	if gotBody.Cert != certPEM || gotBody.Key != keyPEM {
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
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir: dir, CoreVersion: "0.4.28",
	}, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
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

type webhookResult struct {
	Cert    string         `json:"cert"`
	Key     string         `json:"key"`
	Version string         `json:"version"`
	Config  map[string]any `json:"config"`
}

func deployFixture(t *testing.T, factory deployers.ProviderFactoryFunc, url string, results <-chan webhookResult) webhookResult {
	t.Helper()
	instance, err := factory(&deployers.ProviderFactoryOptions{
		ProviderAccessConfig:   map[string]any{"url": url},
		ProviderExtendedConfig: map[string]any{"count": 7},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := instance.Deploy(ctx, "CERT", "KEY"); err != nil {
		t.Fatal(err)
	}
	select {
	case result := <-results:
		if result.Cert != "CERT" || result.Key != "KEY" || result.Config["count"] != float64(7) {
			t.Fatalf("unexpected deployed payload: %+v", result)
		}
		return result
	case <-ctx.Done():
		t.Fatal("webhook did not receive deployment")
	}
	return webhookResult{}
}

func TestE2E_ReloadKeepsDeploymentsAvailableAndUpdatesSchemas(t *testing.T) {
	const providerType = "e2e-lifecycle"
	const accessType = "e2e-lifecycle-access"
	root := t.TempDir()
	starts := filepath.Join(t.TempDir(), "starts")
	fixture := map[string]any{"providerType": providerType, "accessProviderType": accessType, "schemaRevision": "v1", "startupLog": starts}
	dir := setupE2EPlugin(t, root, providerType, accessType, fixture)
	cfg := plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	reloader := pluginhost.NewReloader(cfg, catalog, slog.Default())
	before, err := os.ReadFile(starts)
	if err != nil {
		t.Fatal(err)
	}
	result := reloader.ReloadNow(context.Background())
	if len(result.Errors)+len(result.Added)+len(result.Changed)+len(result.Removed) != 0 {
		t.Fatalf("unchanged reload: %+v", result)
	}
	after, err := os.ReadFile(starts)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("unchanged reload restarted the plugin")
	}
	oldFactory, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan webhookResult, 4)
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseBootstrap := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseBootstrap)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/bootstrap" {
			close(entered)
			select {
			case <-release:
			case <-req.Context().Done():
			}
			return
		}
		var result webhookResult
		if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		results <- result
	}))
	defer srv.Close()
	defer releaseBootstrap()
	fixture["schemaRevision"] = "bad"
	fixture["bootstrapError"] = "fixture cannot bootstrap"
	writeFixtureJSON(t, filepath.Join(dir, "fixture.json"), fixture)
	failed := reloader.ReloadNow(context.Background())
	if len(failed.Errors) == 0 {
		t.Fatal("broken update must fail")
	}
	stillRegistered, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatalf("failed update removed working provider: %v", err)
	}
	if got := deployFixture(t, stillRegistered, srv.URL, results); got.Version != "v1" {
		t.Fatalf("failed update did not preserve v1: %+v", got)
	}
	delete(fixture, "bootstrapError")
	fixture["schemaRevision"] = "v2"
	fixture["accessField"] = "tokenV2"
	fixture["deployField"] = "messageV2"
	fixture["metadataUrl"] = srv.URL + "/bootstrap"
	writeFixtureJSON(t, filepath.Join(dir, "fixture.json"), fixture)
	reloaded := make(chan *pluginhost.ReloadResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go func() { reloaded <- reloader.ReloadNow(ctx) }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("new plugin did not enter bootstrap")
	}
	currentFactory, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		releaseBootstrap()
		t.Fatalf("provider disappeared during bootstrap: %v", err)
	}
	if got := deployFixture(t, currentFactory, srv.URL, results); got.Version != "v1" {
		releaseBootstrap()
		t.Fatalf("pending update changed active deployment: %+v", got)
	}
	releaseBootstrap()
	select {
	case result := <-reloaded:
		if len(result.Errors) != 0 || len(result.Changed) != 1 {
			t.Fatalf("reload failed: %+v", result)
		}
	case <-ctx.Done():
		t.Fatal("reload did not complete")
	}
	svc := providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())
	access, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess)
	if err != nil || !hasColumn(access, "tokenV2") || hasColumn(access, "token") {
		t.Fatalf("old access schema survived: %+v %v", access, err)
	}
	deploy, err := svc.GetByProviderType(context.Background(), providerType, providerschema.CategoryDeploy)
	if err != nil || !hasColumn(deploy, "messageV2") {
		t.Fatalf("deploy schema did not update: %+v %v", deploy, err)
	}
	updatedFactory, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	if got := deployFixture(t, updatedFactory, srv.URL, results); got.Version != "v2" {
		t.Fatalf("updated factory still deploys old version: %+v", got)
	}
	runtime.GC()
	if got := deployFixture(t, oldFactory, srv.URL, results); got.Version != "v1" {
		t.Fatalf("previously acquired factory cannot finish v1 deployment: %+v", got)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	removed := reloader.ReloadNow(context.Background())
	if len(removed.Errors) != 0 || len(removed.Removed) != 1 {
		t.Fatalf("remove failed: %+v", removed)
	}
	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType)); err == nil {
		t.Fatal("removed provider remains registered")
	}
	if env, ok := providerschema.Envelopes.GetEnvelope(accessType, providerschema.CategoryAccess); ok || env != nil {
		t.Fatalf("removed access schema survives: %+v", env)
	}
	if len(catalog.Entries()) != 0 {
		t.Fatalf("removed catalog entry survives: %+v", catalog.Entries())
	}
}

func hasColumn(env *providerschema.Envelope, name string) bool {
	if env == nil {
		return false
	}
	for _, column := range env.Schema.Columns {
		if column.Name == name {
			return true
		}
	}
	return false
}

func TestE2E_RejectsConflictingAndMismatchedIdentities(t *testing.T) {
	builtin, err := deployers.Registries.Get(domain.DeploymentProviderTypeLocal)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, manifestType, metadataType string }{
		{"builtin", string(domain.DeploymentProviderTypeLocal), string(domain.DeploymentProviderTypeLocal)},
		{"mismatch", "e2e-manifest-identity", "e2e-other-identity"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			setupE2EPlugin(t, root, test.manifestType, "e2e-identity-access", map[string]any{"providerType": test.metadataType})
			cfg := plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}
			catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
			t.Cleanup(catalog.Close)
			if len(errs) == 0 || len(catalog.Entries()) != 0 {
				t.Fatalf("invalid identity accepted: %+v %+v", errs, catalog.Entries())
			}
			reloader := pluginhost.NewReloader(cfg, catalog, slog.Default())
			result := reloader.ReloadNow(context.Background())
			if len(result.Errors) == 0 || len(catalog.Entries()) != 0 {
				t.Fatalf("reload accepted invalid identity: %+v", result)
			}
			current, err := deployers.Registries.Get(domain.DeploymentProviderTypeLocal)
			if err != nil || reflect.ValueOf(current).Pointer() != reflect.ValueOf(builtin).Pointer() {
				t.Fatalf("builtin provider replaced: %v", err)
			}
			if test.name == "mismatch" {
				if _, err := deployers.Registries.Get(domain.DeploymentProviderType(test.metadataType)); err == nil {
					t.Fatal("mismatched metadata left a ghost provider")
				}
			}
		})
	}
}

func TestE2E_SameNameAccessAndDeploySchemas(t *testing.T) {
	const providerType = "e2e-same-name"
	root := t.TempDir()
	setupE2EPlugin(t, root, providerType, providerType, map[string]any{"omitMetadataAccessType": true})
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	svc := providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())
	access, err := svc.GetByProviderType(context.Background(), providerType, providerschema.CategoryAccess)
	if err != nil || access.Category != providerschema.CategoryAccess || !hasColumn(access, "url") || hasColumn(access, "count") {
		t.Fatalf("wrong access schema: %+v %v", access, err)
	}
	deploy, err := svc.GetByProviderType(context.Background(), providerType, providerschema.CategoryDeploy)
	if err != nil || deploy.Category != providerschema.CategoryDeploy || !hasColumn(deploy, "count") || hasColumn(deploy, "url") {
		t.Fatalf("wrong deploy schema: %+v %v", deploy, err)
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

func TestE2E_AccessSchemaCanShareBuiltInDeployName(t *testing.T) {
	const providerType = "e2e-builtin-deploy-access"
	accessType := string(domain.DeploymentProviderTypeAliyunCDN)
	root := t.TempDir()
	setupE2EPlugin(t, root, providerType, accessType, nil)
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: root}, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	svc := providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())
	access, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess)
	if err != nil || !hasColumn(access, "url") {
		t.Fatalf("plugin access schema lost to built-in deploy name: %+v %v", access, err)
	}
	deploy, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryDeploy)
	if err != nil || !hasColumn(deploy, "region") || hasColumn(deploy, "url") {
		t.Fatalf("built-in deploy schema changed: %+v %v", deploy, err)
	}
}

func TestE2E_SharedAccessSchemaOwnerHandoff(t *testing.T) {
	const accessType = "e2e-shared-access"
	root := t.TempDir()
	owner := setupE2EPlugin(t, root, "e2e-shared-a", accessType, map[string]any{"accessField": "tokenA"})
	backup := setupE2EPlugin(t, root, "e2e-shared-b", accessType, map[string]any{"accessField": "tokenB"})
	cfg := plugin.PluginConfig{PluginDir: root}
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	svc := providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())
	first, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess)
	if err != nil || !hasColumn(first, "tokenA") {
		t.Fatalf("wrong initial owner: %+v %v", first, err)
	}
	fixturePath := filepath.Join(owner, "fixture.json")
	info, err := os.Stat(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixturePath, []byte(strings.ReplaceAll(string(original), "tokenA", "tokenZ")), info.Mode()); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(fixturePath, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	reloader := pluginhost.NewReloader(cfg, catalog, slog.Default())
	if result := reloader.ReloadNow(context.Background()); len(result.Errors) != 0 {
		t.Fatal(result.Errors)
	}
	updated, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess)
	if err != nil || !hasColumn(updated, "tokenZ") || hasColumn(updated, "tokenA") {
		t.Fatalf("shared owner update stale: %+v %v", updated, err)
	}
	if err := os.RemoveAll(owner); err != nil {
		t.Fatal(err)
	}
	if result := reloader.ReloadNow(context.Background()); len(result.Errors) != 0 {
		t.Fatal(result.Errors)
	}
	remaining, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess)
	if err != nil || !hasColumn(remaining, "tokenB") || hasColumn(remaining, "tokenZ") {
		t.Fatalf("remaining contributor lost schema: %+v %v", remaining, err)
	}
	if err := os.RemoveAll(backup); err != nil {
		t.Fatal(err)
	}
	if result := reloader.ReloadNow(context.Background()); len(result.Errors) != 0 {
		t.Fatal(result.Errors)
	}
	if _, err := svc.GetByProviderType(context.Background(), accessType, providerschema.CategoryAccess); err == nil {
		t.Fatal("schema survived after last contributor was removed")
	}
}
