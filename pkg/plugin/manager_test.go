package plugin

import (
	"bytes"
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
	"time"
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

func TestManager_Deploy_RedactsReturnedError(t *testing.T) {
	defer withFakeEnv("error-echo", "error-echo")()
	dp := discoveredFake(t, "error-echo")
	const token = "fixture-returned-error-token"
	const key = "-----BEGIN PRIVATE KEY-----\nZmFrZS1wcml2YXRlLWtleS1ib2R5\n-----END PRIVATE KEY-----"
	request := &DeployRequest{AccessConfigJSON: `{"apiToken":"` + token + `"}`, PrivateKeyPEM: key}
	_, err := NewManager(PluginConfig{}, nil).Deploy(context.Background(), dp, request, nil)
	if err == nil || !strings.Contains(err.Error(), "remote rejected") {
		t.Fatalf("expected remote rejection, got %v", err)
	}
	for _, secret := range []string{token, key, "ZmFrZS1wcml2YXRlLWtleS1ib2R5"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatal("plugin error exposed deployment credentials")
		}
	}
}

func TestSyncBuffer_ConcurrentWriteAndTail(t *testing.T) {
	var b syncBuffer

	const line = "stderr line\n"
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b.Write([]byte(line))
			}
		}()
	}
	for i := 0; i < 100; i++ {
		b.tail(stderrTailLimit)
	}
	wg.Wait()

	if got, want := len(b.tail(stderrTailLimit)), 4*100*len(line); got != want {
		t.Fatalf("tail length: got %d, want %d", got, want)
	}

	for i := 0; i < 3*stderrTailLimit/len(line); i++ {
		b.Write([]byte(line))
	}
	if b.buf.Len() > 2*stderrTailLimit {
		t.Fatalf("compacted buffer length: got %d, want <= %d", b.buf.Len(), 2*stderrTailLimit)
	}
	tail := b.tail(stderrTailLimit)
	if len(tail) != stderrTailLimit {
		t.Fatalf("capped tail length: got %d, want %d", len(tail), stderrTailLimit)
	}
	if !bytes.HasSuffix(tail, []byte(line)) {
		t.Fatal("capped tail should end with a complete line")
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

func discoveredFakeCapability(t *testing.T, providerType, capability, behavior string) *DiscoveredPlugin {
	t.Helper()
	bin := fakeBinary(t)
	dir := t.TempDir()
	manifest := manifestJSON(func(m map[string]any) {
		m["provider_type"] = providerType
		m["access_provider_type"] = providerType
		m["binary"] = "fakeplugin"
		m["capability"] = capability
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
	t.Setenv("FAKEPLUGIN_PROVIDER_TYPE", providerType)
	t.Setenv("FAKEPLUGIN_CAPABILITY", capability)
	t.Setenv("FAKEPLUGIN_BEHAVIOR", behavior)
	state := filepath.Join(t.TempDir(), "state.txt")
	t.Setenv("FAKEPLUGIN_STATE_FILE", state)
	return dp
}

func readState(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func TestManager_DnsHeldLifecycle_SameProcessAcrossPresentAndCleanUp(t *testing.T) {
	dp := discoveredFakeCapability(t, "dns-demo", CapabilityDns, "ok")
	state := os.Getenv("FAKEPLUGIN_STATE_FILE")

	mgr := NewManager(PluginConfig{}, nil)
	held, err := mgr.StartDnsProvider(context.Background(), dp)
	if err != nil {
		t.Fatalf("start dns provider: %v", err)
	}
	req := &DnsRequest{Domain: "example.test", Token: "tok", KeyAuth: "ka"}
	if err := held.Present(context.Background(), req); err != nil {
		t.Fatalf("present: %v", err)
	}
	if err := held.CleanUp(context.Background(), req); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	held.Close()
	events := readState(t, state)
	if len(events) != 2 || events[0] != "present" || events[1] != "cleanup" {
		t.Fatalf("expected present then cleanup in one process, got %v", events)
	}
}

func TestManager_DnsHeldLifecycle_ProcessDeathFailsCleanUp(t *testing.T) {
	dp := discoveredFakeCapability(t, "dns-crash", CapabilityDns, "present-crash")
	state := os.Getenv("FAKEPLUGIN_STATE_FILE")

	mgr := NewManager(PluginConfig{}, nil)
	held, err := mgr.StartDnsProvider(context.Background(), dp)
	if err != nil {
		t.Fatalf("start dns provider: %v", err)
	}
	req := &DnsRequest{Domain: "example.test"}
	if err := held.Present(context.Background(), req); err != nil {
		t.Fatalf("present: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
	err = held.CleanUp(context.Background(), req)
	if err == nil {
		t.Fatal("expected cleanup failure after process death")
	}
	var crashed *ErrPluginCrashed
	if !errors.As(err, &crashed) {
		t.Fatalf("expected ErrPluginCrashed, got %v", err)
	}
	events := readState(t, state)
	if len(events) != 1 || events[0] != "present" {
		t.Fatalf("expected no respawn cleanup, got %v", events)
	}
}

func TestManager_Dns_PresentErrorReleasesProcess(t *testing.T) {
	dp := discoveredFakeCapability(t, "dns-err", CapabilityDns, "present-error")
	state := os.Getenv("FAKEPLUGIN_STATE_FILE")

	mgr := NewManager(PluginConfig{}, nil)
	held, err := mgr.StartDnsProvider(context.Background(), dp)
	if err != nil {
		t.Fatalf("start dns provider: %v", err)
	}
	req := &DnsRequest{Domain: "example.test"}
	if err := held.Present(context.Background(), req); err == nil {
		t.Fatal("expected present error")
	}
	err = held.CleanUp(context.Background(), req)
	if err == nil {
		t.Fatal("expected cleanup error after release")
	}
	events := readState(t, state)
	if len(events) != 1 || events[0] != "present:error" {
		t.Fatalf("expected present error only, got %v", events)
	}
}

func TestManager_Notify_RoundTrip(t *testing.T) {
	dp := discoveredFakeCapability(t, "notify-demo", CapabilityNotify, "ok")
	state := os.Getenv("FAKEPLUGIN_STATE_FILE")

	mgr := NewManager(PluginConfig{}, nil)
	res, err := mgr.Notify(context.Background(), dp, &NotifyRequest{Subject: "hello", Message: "world"})
	if err != nil {
		t.Fatalf("notify: %v", err)
	}
	if res == nil {
		t.Fatal("nil notify result")
	}
	events := readState(t, state)
	if len(events) != 1 || events[0] != "notify:hello|world" {
		t.Fatalf("unexpected notify events: %v", events)
	}
}

func TestManager_Notify_ErrorPropagates(t *testing.T) {
	dp := discoveredFakeCapability(t, "notify-err", CapabilityNotify, "notify-error")

	mgr := NewManager(PluginConfig{}, nil)
	_, err := mgr.Notify(context.Background(), dp, &NotifyRequest{Subject: "x", Message: "y"})
	if err == nil || !strings.Contains(err.Error(), "fake notify failure") {
		t.Fatalf("expected notify failure, got %v", err)
	}
}

func TestManager_Deploy_RejectsWrongCapability(t *testing.T) {
	dp := discoveredFakeCapability(t, "dns-only", CapabilityDns, "ok")

	mgr := NewManager(PluginConfig{}, nil)
	_, err := mgr.Deploy(context.Background(), dp, sampleDeployReq(), nil)
	if err == nil {
		t.Fatal("expected deploy on dns capability to fail")
	}
	var inc *ErrPluginIncompatible
	if !errors.As(err, &inc) {
		t.Fatalf("expected ErrPluginIncompatible, got %v", err)
	}
}
