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

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/plugin"
)

var (
	hostFakeOnce sync.Once
	hostFakeBin  string
	hostFakeErr  error
)

func buildFakeBinary(t *testing.T) string {
	t.Helper()
	hostFakeOnce.Do(func() {
		out := filepath.Join(os.TempDir(), "certimate-fakeplugin-pluginhost-test")
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		cmd := exec.Command("go", "build", "-tags=fakeplugin", "-o", out, "github.com/certimate-go/certimate/pkg/plugin/testplugin")
		if buildOut, err := cmd.CombinedOutput(); err != nil {
			hostFakeErr = errors.New(strings.TrimSpace(string(buildOut)) + " | " + err.Error())
			return
		}
		hostFakeBin = out
	})
	if hostFakeErr != nil {
		t.Skipf("fakeplugin build failed: %v", hostFakeErr)
	}
	return hostFakeBin
}

func setupPluginDir(t *testing.T, providerType, accessType string) (string, *plugin.DiscoveredPlugin) {
	t.Helper()
	bin := buildFakeBinary(t)
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, providerType)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"version":              "0.1.0",
		"provider_type":        providerType,
		"access_provider_type": accessType,
		"display_name_key":     "plugin." + providerType + ".name",
		"deploy_category":      "other",
		"protocol_version":     plugin.ProtocolVersion,
		"binary":               "fakeplugin",
	}
	mb := jsonMarshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.json"), mb, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bin, filepath.Join(pluginDir, "fakeplugin")); err != nil {
		t.Fatal(err)
	}
	return dir, &plugin.DiscoveredPlugin{Manifest: &plugin.Manifest{
		ProviderType:       providerType,
		AccessProviderType: accessType,
		Binary:             "fakeplugin",
	}, Dir: pluginDir}
}

func jsonMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func setFakeEnv(providerType, behavior string) func() {
	os.Setenv("FAKEPLUGIN_PROVIDER_TYPE", providerType)
	os.Setenv("FAKEPLUGIN_ACCESS_TYPE", providerType+"-access")
	os.Setenv("FAKEPLUGIN_BEHAVIOR", behavior)
	return func() {
		os.Unsetenv("FAKEPLUGIN_PROVIDER_TYPE")
		os.Unsetenv("FAKEPLUGIN_ACCESS_TYPE")
		os.Unsetenv("FAKEPLUGIN_BEHAVIOR")
	}
}

func TestScanAndRegister_RegistersFactorySchemaAndCatalog(t *testing.T) {
	deployType := "hosttest-reg-" + t.Name()
	accessType := deployType + "-access"
	defer setFakeEnv(deployType, "ok")()
	dir, _ := setupPluginDir(t, deployType, accessType)

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir:   dir,
		CoreVersion: "0.4.28",
	}, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %+v", errs)
	}

	entries := catalog.Entries()
	if len(entries) != 1 || entries[0].ProviderType != deployType {
		t.Fatalf("catalog entry missing: %+v", entries)
	}

	factory, err := deployers.Registries.Get(domain.DeploymentProviderType(deployType))
	if err != nil {
		t.Fatalf("deployer not registered: %v", err)
	}
	deployer, err := factory(&deployers.ProviderFactoryOptions{
		ProviderAccessConfig:   map[string]any{"url": "https://example.com", "secret": "tok"},
		ProviderExtendedConfig: map[string]any{"method": "POST"},
	})
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	deployer.SetLogger(slog.Default())
	res, err := deployer.Deploy(context.Background(), "CERT", "KEY")
	if err != nil {
		t.Fatalf("deploy via adapter: %v", err)
	}
	if res == nil {
		t.Fatal("nil deploy result")
	}

	if env, ok := providerschema.Envelopes.GetEnvelope(deployType); !ok || env == nil {
		t.Fatal("deploy schema envelope not registered")
	}
	if env, ok := providerschema.Envelopes.GetEnvelope(accessType); !ok || env == nil {
		t.Fatal("access schema envelope not registered")
	}
}

func TestScanAndRegister_BrokenPluginSkipped_CoreContinues(t *testing.T) {
	goodType := "hosttest-broken-good"
	defer setFakeEnv(goodType, "ok")()
	dir, _ := setupPluginDir(t, goodType, goodType+"-access")

	badDir := filepath.Join(dir, "hosttest-broken-bad")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mb := jsonMarshal(map[string]any{
		"provider_type":        "hosttest-broken-bad",
		"access_provider_type": "hosttest-broken-bad",
		"protocol_version":     plugin.ProtocolVersion,
		"binary":               "missing-binary",
	})
	if err := os.WriteFile(filepath.Join(badDir, "manifest.json"), mb, 0o644); err != nil {
		t.Fatal(err)
	}

	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir:   dir,
		CoreVersion: "0.4.28",
	}, slog.Default())
	if len(errs) == 0 {
		t.Fatal("expected at least one error for the broken plugin")
	}
	if len(catalog.Entries()) != 1 || catalog.Entries()[0].ProviderType != goodType {
		t.Fatalf("good plugin should still register: %+v", catalog.Entries())
	}
}

func TestScanAndRegister_EmptyDir_NoOp(t *testing.T) {
	dir := t.TempDir()
	catalog, errs := ScanAndRegister(context.Background(), plugin.PluginConfig{
		PluginDir:   dir,
		CoreVersion: "0.4.28",
	}, slog.Default())
	if len(errs) != 0 || len(catalog.Entries()) != 0 {
		t.Fatalf("empty dir should be no-op: errs=%+v entries=%+v", errs, catalog.Entries())
	}
}

func TestScanAndRegister_BuiltinProvidersStillRegistered(t *testing.T) {
	if _, err := deployers.Registries.Get(domain.DeploymentProviderTypeLocal); err != nil {
		t.Fatalf("builtin LOCAL deployer missing after plugin scan wiring: %v", err)
	}
}
