package pluginhost

import (
	"context"
	"encoding/json"
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

	"github.com/certimate-go/certimate/internal/accesscheck"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func draftTestReq(provider string, config map[string]any) *dtos.AccessTestReq {
	return &dtos.AccessTestReq{Source: dtos.AccessTestSourceDraft, Provider: provider, Config: config}
}

func TestAccessCheck_OwnerCheckerRegisteredAndCheckPasses(t *testing.T) {
	deployType := "acheck-owner-ok"
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "ok")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: dir, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	entry, ok := accesscheck.DefaultRegistry().Get(accessType)
	if !ok {
		t.Fatal("owner plugin must register an access check entry")
	}
	if entry.Source != accesscheck.SourcePlugin || entry.OwnerProviderType != deployType || entry.Checker == nil {
		t.Fatalf("registration mismatch: %+v", entry)
	}
	if !accesscheck.DefaultRegistry().IsKnownProvider(accessType) {
		t.Fatal("catalog-known access type must resolve as known provider")
	}

	svc := accesscheck.NewService(nil)
	caps := svc.TestCapabilities(accessType)
	if !caps.Supported || caps.Source != accesscheck.SourcePlugin || caps.OwnerProviderType != deployType {
		t.Fatalf("capabilities mismatch: %+v", caps)
	}
	if len(caps.Scope) != 2 || caps.Scope[0] != "connect" || caps.Scope[1] != "auth" {
		t.Fatalf("capabilities scope mismatch: %+v", caps.Scope)
	}

	resp, err := svc.Test(context.Background(), "tester", draftTestReq(accessType, map[string]any{"apiKey": "super-secret-value", "url": "https://example.com"}))
	if err != nil {
		t.Fatalf("draft test through real plugin binary: %v", err)
	}
	if resp.Status != dtos.AccessTestStatusPassed || len(resp.Items) != 2 {
		t.Fatalf("check result mismatch: %+v", resp)
	}

	catalog.Close()
	if _, still := accesscheck.DefaultRegistry().Get(accessType); still {
		t.Fatal("closing the catalog must drop the access check registration")
	}
}

func TestAccessCheck_PluginErrorMapsToUnavailableAndRedactsSecret(t *testing.T) {
	deployType := "acheck-err-echo"
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "check-error-echo")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: dir, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	svc := accesscheck.NewService(nil)
	resp, err := svc.Test(context.Background(), "tester", draftTestReq(accessType, map[string]any{"apiKey": "super-secret-value"}))
	if err != nil {
		t.Fatalf("plugin-side failure must surface as a result, got error %v", err)
	}
	if resp.Status != dtos.AccessTestStatusUnavailable || resp.ReasonCode != string(dtos.AccessTestReasonCodePluginUnavailable) {
		t.Fatalf("status mismatch: %+v", resp)
	}
	joined := strings.Join(resp.Diagnostics, "\n")
	if strings.Contains(joined, "super-secret-value") {
		t.Fatalf("secret leaked through plugin error: %s", joined)
	}
	if !strings.Contains(joined, "[REDACTED]") {
		t.Fatalf("secret must be redacted, got: %s", joined)
	}
}

func TestAccessCheck_SuccessDiagnosticsRedacted(t *testing.T) {
	deployType := "acheck-diag-echo"
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "check-diagnostic-echo")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: dir, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	svc := accesscheck.NewService(nil)
	resp, err := svc.Test(context.Background(), "tester", draftTestReq(accessType, map[string]any{"apiKey": "super-secret-value"}))
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.Status != dtos.AccessTestStatusPassed {
		t.Fatalf("status mismatch: %+v", resp)
	}
	joined := strings.Join(resp.Diagnostics, "\n")
	if strings.Contains(joined, "super-secret-value") {
		t.Fatalf("secret leaked through result diagnostics: %s", joined)
	}
	if !strings.Contains(joined, "[REDACTED]") {
		t.Fatalf("diagnostics must be redacted, got: %s", joined)
	}
}

func TestAccessCheck_DeadlinePropagatesToSubprocess(t *testing.T) {
	deployType := "acheck-block"
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "check-block")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: dir, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	svc := accesscheck.NewService(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	startedAt := time.Now()
	resp, err := svc.Test(ctx, "tester", draftTestReq(accessType, map[string]any{"apiKey": "super-secret-value"}))
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	elapsed := time.Since(startedAt)
	if elapsed > 10*time.Second {
		t.Fatalf("caller deadline must abort the subprocess call, took %s", elapsed)
	}
	if resp.Status != dtos.AccessTestStatusInconclusive || resp.ReasonCode != string(dtos.AccessTestReasonCodeTimeout) {
		t.Fatalf("timeout mapping mismatch: %+v", resp)
	}
}

func TestAccessCheck_CrashDuringCheckRedactsStderr(t *testing.T) {
	deployType := "acheck-crash"
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "check-crash")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: dir, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	entry, ok := accesscheck.DefaultRegistry().Get(accessType)
	if !ok || entry.Checker == nil {
		t.Fatalf("registration mismatch: %+v", entry)
	}
	resp, err := entry.Checker.Check(context.Background(), map[string]any{"apiKey": "super-secret-value"})
	if err != nil {
		t.Fatalf("crashed check must surface as unavailable result, got error %v", err)
	}
	joined := strings.Join(resp.Diagnostics, "\n")
	if strings.Contains(joined, "super-secret-value") {
		t.Fatalf("secret leaked through crash diagnostics: %s", joined)
	}
	if !strings.Contains(joined, "[REDACTED]") {
		t.Fatalf("crash diagnostics must be redacted, got: %s", joined)
	}
}

func TestAccessCheck_OwnerWithoutInterfaceRegistersNoChecker(t *testing.T) {
	provider := "acheck-noif"
	p := &registeredPlugin{
		ownsAccess: true,
		accessType: provider,
		accessCheck: &accessCheckState{
			accessType:        provider,
			ownerProviderType: provider,
		},
	}
	reconcileAccessCheckRegistration(p)
	t.Cleanup(func() { accesscheck.DefaultRegistry().Unregister(provider) })

	entry, ok := accesscheck.DefaultRegistry().Get(provider)
	if !ok || entry.Checker != nil || entry.UnavailableReason != dtos.AccessTestReasonCodeNoChecker {
		t.Fatalf("registration mismatch: %+v", entry)
	}
	caps := accesscheck.NewService(nil).TestCapabilities(provider)
	if caps.Supported || caps.ReasonCode != string(dtos.AccessTestReasonCodeNoChecker) {
		t.Fatalf("capabilities mismatch: %+v", caps)
	}
}

func TestAccessCheck_ReuserNeverDisplacesOwner(t *testing.T) {
	bin := buildAccessCheckFixture(t)
	root := t.TempDir()
	writeAccessCheckFixtureDir(t, filepath.Join(root, "acheck-or-owner"), map[string]any{
		"providerType":       "acheck-or-owner",
		"accessProviderType": "acheck-or",
		"schemaRevision":     "v1",
	}, bin)
	writeAccessCheckFixtureDir(t, filepath.Join(root, "acheck-or-reuser"), map[string]any{
		"providerType":       "acheck-or-reuser",
		"accessProviderType": "acheck-or",
		"omitAccessSchema":   true,
		"schemaRevision":     "v1",
	}, bin)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("scan errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)

	entry, ok := accesscheck.DefaultRegistry().Get("acheck-or")
	if !ok {
		t.Fatal("access owner must register the check entry")
	}
	if entry.OwnerProviderType != "acheck-or-owner" {
		t.Fatalf("reuser must never displace the owner as check provider: %+v", entry)
	}
	resp, err := accesscheck.NewService(nil).Test(context.Background(), "tester", draftTestReq("acheck-or", map[string]any{"url": "https://example.com", "token": "secret-token-value"}))
	if err != nil {
		t.Fatalf("draft test routed to owner plugin: %v", err)
	}
	if resp.Status != dtos.AccessTestStatusPassed {
		t.Fatalf("owner check result mismatch: %+v", resp)
	}
}

func TestAccessCheck_StaleProtocolOwnerRegistersProtocolOld(t *testing.T) {
	root := t.TempDir()
	staleDir := filepath.Join(root, "acheck-old")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staleDir, "manifest.json"), jsonMarshal(map[string]any{
		"provider_type":        "acheck-old",
		"access_provider_type": "acheck-old",
		"protocol_version":     plugin.ProtocolVersion - 1,
		"capability":           plugin.CapabilityDeploy,
		"binary":               "never-checked",
	}), 0o644); err != nil {
		t.Fatal(err)
	}

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}, slog.Default())
	if len(errs) == 0 {
		t.Fatal("stale plugin must be reported as a scan error")
	}
	t.Cleanup(catalog.Close)
	t.Cleanup(func() { accesscheck.DefaultRegistry().Unregister("acheck-old") })

	entry, ok := accesscheck.DefaultRegistry().Get("acheck-old")
	if !ok || entry.Checker != nil || entry.UnavailableReason != dtos.AccessTestReasonCodePluginProtocolOld {
		t.Fatalf("registration mismatch: %+v", entry)
	}
	caps := accesscheck.NewService(nil).TestCapabilities("acheck-old")
	if caps.Supported || caps.ReasonCode != string(dtos.AccessTestReasonCodePluginProtocolOld) {
		t.Fatalf("capabilities mismatch: %+v", caps)
	}
}

var (
	accessCheckFixtureOnce sync.Once
	accessCheckFixtureBin  string
	accessCheckFixtureErr  error
)

func buildAccessCheckFixture(t *testing.T) string {
	t.Helper()
	accessCheckFixtureOnce.Do(func() {
		out := filepath.Join(os.TempDir(), "certimate-accesscheck-testplugin")
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", out, "github.com/certimate-go/certimate/internal/pluginhost/testplugin")
		if buildOut, err := cmd.CombinedOutput(); err != nil {
			accessCheckFixtureErr = errors.New(strings.TrimSpace(string(buildOut)) + " | " + err.Error())
			return
		}
		accessCheckFixtureBin = out
	})
	if accessCheckFixtureErr != nil {
		t.Skipf("access check fixture build failed: %v", accessCheckFixtureErr)
	}
	return accessCheckFixtureBin
}

func writeAccessCheckFixtureDir(t *testing.T, dir string, fixture map[string]any, bin string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerType, _ := fixture["providerType"].(string)
	accessType, _ := fixture["accessProviderType"].(string)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), jsonMarshal(map[string]any{
		"version":              "0.1.0",
		"provider_type":        providerType,
		"access_provider_type": accessType,
		"capability":           plugin.CapabilityDeploy,
		"display_name_key":     "plugin.e2e.name",
		"category":             "other",
		"protocol_version":     plugin.ProtocolVersion,
		"binary":               "fixture-bin",
	}), 0o644); err != nil {
		t.Fatal(err)
	}
	writeJSONFile(t, filepath.Join(dir, "fixture.json"), fixture)
	data, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fixture-bin"), data, 0o755); err != nil {
		t.Fatal(err)
	}
}
