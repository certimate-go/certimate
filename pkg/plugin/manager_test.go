package plugin

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	fakeBinaryOnce sync.Once
	fakeBinaryPath string
	fakeBinaryErr  error
)

func fakeBinary(t *testing.T) string {
	t.Helper()
	fakeBinaryOnce.Do(func() {
		out := filepath.Join(os.TempDir(), "certimate-fakeplugin-test")
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		cmd := exec.Command("go", "build", "-tags=fakeplugin", "-o", out, "./testplugin")
		cmd.Dir = "."
		if buildOut, err := cmd.CombinedOutput(); err != nil {
			fakeBinaryErr = errors.New(strings.TrimSpace(string(buildOut)) + " | " + err.Error())
			return
		}
		fakeBinaryPath = out
	})
	if fakeBinaryErr != nil {
		t.Skipf("fakeplugin build failed: %v", fakeBinaryErr)
	}
	return fakeBinaryPath
}

func discoveredFake(t *testing.T, providerType string) *DiscoveredPlugin {
	t.Helper()
	bin := fakeBinary(t)
	dir := t.TempDir()
	manifest := manifestJSON(func(m map[string]any) {
		m["provider_type"] = providerType
		m["access_provider_type"] = providerType
		m["binary"] = "fakeplugin"
	})
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "fakeplugin")
	if err := os.Symlink(bin, link); err != nil {
		t.Fatal(err)
	}
	dp, ferr := discoverOne(dir, "0.1.0")
	if ferr != nil {
		t.Fatalf("discoverOne: %v", ferr.Err)
	}
	return dp
}

func withFakeEnv(providerType, behavior string) func() {
	os.Setenv("FAKEPLUGIN_PROVIDER_TYPE", providerType)
	os.Setenv("FAKEPLUGIN_BEHAVIOR", behavior)
	return func() {
		os.Unsetenv("FAKEPLUGIN_PROVIDER_TYPE")
		os.Unsetenv("FAKEPLUGIN_BEHAVIOR")
	}
}

func TestManager_Bootstrap_FetchesMetadataAndSchema(t *testing.T) {
	defer withFakeEnv("bootstrap-demo", "ok")()
	dp := discoveredFake(t, "bootstrap-demo")

	mgr := NewManager(PluginConfig{}, nil)
	meta, schema, err := mgr.Bootstrap(context.Background(), dp)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if meta.ProviderType != "bootstrap-demo" || meta.ProtocolVersion != ProtocolVersion {
		t.Fatalf("metadata mismatch: %+v", meta)
	}
	if schema == nil || len(schema.DeploySchemaJSON) == 0 {
		t.Fatalf("schema not fetched: %+v", schema)
	}
	if schema.I18n["en"]["plugin.fake.name"] == "" {
		t.Fatalf("i18n not fetched: %+v", schema.I18n)
	}
}

func TestManager_Deploy_OnDemandLifecycle(t *testing.T) {
	defer withFakeEnv("deploy-demo", "ok")()
	dp := discoveredFake(t, "deploy-demo")

	mgr := NewManager(PluginConfig{}, nil)
	res1, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), nil)
	if err != nil {
		t.Fatalf("first deploy: %v", err)
	}
	if res1.ExtendedDataJSON == "" {
		t.Fatal("empty deploy result")
	}
	res2, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), nil)
	if err != nil {
		t.Fatalf("second deploy (fresh client): %v", err)
	}
	if res2 == nil {
		t.Fatal("nil second result")
	}
}

func TestManager_Deploy_PluginConfigError_Mapped(t *testing.T) {
	defer withFakeEnv("cfgerr-demo", "configerror")()
	dp := discoveredFake(t, "cfgerr-demo")

	mgr := NewManager(PluginConfig{}, nil)
	_, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), nil)
	if err == nil {
		t.Fatal("expected config error")
	}
	if !strings.Contains(err.Error(), "bad config") {
		t.Fatalf("expected config error message preserved, got %q", err.Error())
	}
}

func TestManager_Deploy_CrashIsolated_ReturnsErrPluginCrashed(t *testing.T) {
	defer withFakeEnv("crash-demo", "crash")()
	dp := discoveredFake(t, "crash-demo")

	mgr := NewManager(PluginConfig{}, nil)
	_, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), nil)
	if err == nil {
		t.Fatal("expected crash error")
	}
	var crashed *ErrPluginCrashed
	if !errors.As(err, &crashed) {
		t.Fatalf("expected ErrPluginCrashed, got %T: %v", err, err)
	}
	if crashed.ProviderType != "crash-demo" {
		t.Fatalf("crash error wrong provider: %q", crashed.ProviderType)
	}
	if crashed.StderrTail == "" {
		t.Fatal("expected non-empty stderr tail")
	}
}

func TestManager_Deploy_CrashRedactsCredentials(t *testing.T) {
	defer withFakeEnv("crash-redact", "crash")()
	dp := discoveredFake(t, "crash-redact")

	const secret = "supersecrettoken1234"
	req := &DeployRequest{
		LogLevel:           "INFO",
		AccessConfigJSON:   `{"url":"https://example.com","secret":"` + secret + `"}`,
		ExtendedConfigJSON: `{"path":"/x"}`,
		CertificatePEM:     "CERT",
		PrivateKeyPEM:      "KEY",
	}

	mgr := NewManager(PluginConfig{}, nil)
	_, err := mgr.Deploy(context.Background(), dp, req, nil)
	var crashed *ErrPluginCrashed
	if !errors.As(err, &crashed) {
		t.Fatalf("expected ErrPluginCrashed, got %T: %v", err, err)
	}
	if strings.Contains(crashed.StderrTail, secret) {
		t.Fatalf("secret leaked into crash stderr tail:\n%s", crashed.StderrTail)
	}
}

func TestRedactorFor_RedactsPEM(t *testing.T) {
	body := strings.Repeat("A", 64)
	key := "-----BEGIN PRIVATE KEY-----\n" + body + "\n-----END PRIVATE KEY-----"
	redact := redactorFor(&DeployRequest{PrivateKeyPEM: key})

	fullOut := redact("leaked: " + key)
	if strings.Contains(fullOut, body) {
		t.Fatalf("private key body not redacted when full PEM logged: %q", fullOut)
	}
	bodyOut := redact("leaked body: " + body)
	if strings.Contains(bodyOut, body) {
		t.Fatalf("private key body line not redacted: %q", bodyOut)
	}
}

func TestManager_Deploy_ForwardsPluginLogs(t *testing.T) {
	defer withFakeEnv("forward-demo", "ok")()
	dp := discoveredFake(t, "forward-demo")

	mgr := NewManager(PluginConfig{}, nil)
	cap := &captureHandler{}
	_, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), slog.New(cap))
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	found := false
	for _, r := range cap.records {
		if r.Message == "fakeplugin deploy starting" {
			found = true
		}
	}
	if !found {
		t.Fatalf("plugin log not forwarded to sink: %+v", cap.records)
	}
}

func sampleDeployReq() *DeployRequest {
	return &DeployRequest{
		LogLevel:           "INFO",
		AccessConfigJSON:   `{"url":"https://example.com","secret":"tok"}`,
		ExtendedConfigJSON: `{"method":"POST"}`,
		CertificatePEM:     "CERT",
		PrivateKeyPEM:      "KEY",
	}
}
