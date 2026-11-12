package pluginhost_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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

type marketReleaseFixture struct {
	manifest plugin.MarketManifest
	archive  []byte
}

func buildMarketRelease(t *testing.T, providerType, version, bootstrapError string) marketReleaseFixture {
	t.Helper()
	binary, err := os.ReadFile(buildPilot(t))
	if err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{
		ProviderType: providerType, AccessProviderType: providerType + "-access", Version: version,
		Binary: "webhook-deployer", ProtocolVersion: plugin.ProtocolVersion,
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	configJSON, err := json.Marshal(map[string]any{
		"providerType": providerType, "accessProviderType": manifest.AccessProviderType,
		"schemaRevision": version, "bootstrapError": bootstrapError,
		"accessField": "token-" + version, "deployField": "message-" + version,
	})
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range []struct {
		name string
		data []byte
	}{
		{manifest.Binary, binary}, {"manifest.json", manifestJSON}, {"fixture.json", configJSON},
	} {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Store}
		header.SetMode(0o644)
		output, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := output.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buffer.Bytes()
	sum := sha256.Sum256(archive)
	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	return marketReleaseFixture{
		manifest: plugin.MarketManifest{Manifest: manifest, Release: &plugin.Release{
			Repo: "certimate-go/plugins", Tag: providerType + "/v" + version,
			Assets: map[string]string{key: "plugin.zip"}, Checksums: map[string]string{key: hex.EncodeToString(sum[:])},
		}},
		archive: archive,
	}
}

func marketJobResult(t *testing.T, job *pluginhost.InstallJob) pluginhost.JobStatus {
	t.Helper()
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		status := job.Status()
		if status.State.Terminal() {
			return status
		}
		select {
		case <-deadline.C:
			t.Fatalf("market job did not finish: %+v", status)
		case <-ticker.C:
		}
	}
}

func TestE2E_MarketInstallUpgradeRollbackAndDelete(t *testing.T) {
	const providerType = "e2e-market-lifecycle"
	release := buildMarketRelease(t, providerType, "1.0.0", "")
	var releaseMu sync.RWMutex
	results := make(chan webhookResult, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/deploy" {
			var result webhookResult
			if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			results <- result
			return
		}
		releaseMu.RLock()
		defer releaseMu.RUnlock()
		if req.URL.Path == "/index.json" {
			if err := json.NewEncoder(w).Encode(map[string]any{"plugins": []plugin.MarketManifest{release.manifest}}); err != nil {
				t.Error(err)
			}
			return
		}
		if strings.Contains(req.URL.Path, "/releases/download/") {
			_, _ = w.Write(release.archive)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	root := t.TempDir()
	catalog := pluginhost.NewCatalog()
	t.Cleanup(catalog.Close)
	previousReloader := pluginhost.GlobalReloader()
	pluginhost.SetGlobalReloader(pluginhost.NewReloader(plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}, catalog, slog.Default()))
	t.Cleanup(func() { pluginhost.SetGlobalReloader(previousReloader) })
	market := pluginhost.NewMarketService(pluginhost.MarketConfig{PluginDir: root, IndexURL: server.URL + "/index.json", DownloadMirror: server.URL, CacheTTL: time.Nanosecond})
	job, err := market.Install(context.Background(), providerType)
	if err != nil {
		t.Fatal(err)
	}
	if status := marketJobResult(t, job); status.State != pluginhost.JobInstalled {
		t.Fatalf("installation failed: %+v", status)
	}
	oldFactory, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	if result := deployFixture(t, oldFactory, server.URL+"/deploy", results); result.Version != "1.0.0" {
		t.Fatalf("wrong installed version: %+v", result)
	}
	originalFiles := make(map[string][sha256.Size]byte)
	for _, name := range []string{"manifest.json", "fixture.json", "webhook-deployer", ".market.json"} {
		data, err := os.ReadFile(filepath.Join(root, providerType, name))
		if err != nil {
			t.Fatal(err)
		}
		originalFiles[name] = sha256.Sum256(data)
	}
	schemas := providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())
	originalCatalog, err := json.Marshal(catalog.Entries())
	if err != nil {
		t.Fatal(err)
	}
	broken := buildMarketRelease(t, providerType, "2.0.0", "release fails bootstrap")
	releaseMu.Lock()
	release = broken
	releaseMu.Unlock()
	if _, err := market.Update(context.Background(), providerType); err == nil {
		t.Fatal("broken release update succeeded")
	}
	for name, originalSum := range originalFiles {
		data, err := os.ReadFile(filepath.Join(root, providerType, name))
		if err != nil || sha256.Sum256(data) != originalSum {
			t.Fatalf("failed update changed installed %s: %v", name, err)
		}
	}
	afterFailure, err := json.Marshal(catalog.Entries())
	if err != nil || !bytes.Equal(originalCatalog, afterFailure) {
		t.Fatalf("failed update changed catalog: %s %v", afterFailure, err)
	}
	preservedSchema, err := schemas.GetByProviderType(context.Background(), providerType+"-access", providerschema.CategoryAccess)
	if err != nil || !hasColumn(preservedSchema, "token-1.0.0") || hasColumn(preservedSchema, "token-2.0.0") {
		t.Fatalf("failed update changed access schema: %+v %v", preservedSchema, err)
	}
	metadata, err := plugin.ReadMarketMeta(root, providerType)
	if err != nil || metadata == nil || metadata.InstalledVersion != "1.0.0" {
		t.Fatalf("failed update changed installed version: %+v %v", metadata, err)
	}
	retained, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	if result := deployFixture(t, retained, server.URL+"/deploy", results); result.Version != "1.0.0" {
		t.Fatalf("failed update replaced working deployment: %+v", result)
	}
	valid := buildMarketRelease(t, providerType, "2.0.0", "")
	releaseMu.Lock()
	release = valid
	releaseMu.Unlock()
	result, err := market.Update(context.Background(), providerType)
	if err != nil || result == nil || len(result.Errors) != 0 {
		t.Fatalf("valid update failed: %+v %v", result, err)
	}
	updated, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	if result := deployFixture(t, updated, server.URL+"/deploy", results); result.Version != "2.0.0" {
		t.Fatalf("updated binary did not deploy: %+v", result)
	}
	access, err := schemas.GetByProviderType(context.Background(), providerType+"-access", providerschema.CategoryAccess)
	if err != nil || !hasColumn(access, "token-2.0.0") || hasColumn(access, "token-1.0.0") {
		t.Fatalf("updated access schema stale: %+v %v", access, err)
	}
	if _, err := market.Delete(context.Background(), providerType); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, providerType)); !os.IsNotExist(err) {
		t.Fatalf("deleted plugin directory remains: %v", err)
	}
	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType)); err == nil {
		t.Fatal("deleted provider remains available for new work")
	}
	if len(catalog.Entries()) != 0 {
		t.Fatalf("deleted plugin remains in catalog: %+v", catalog.Entries())
	}
	if _, ok := providerschema.Envelopes.GetEnvelope(providerType+"-access", providerschema.CategoryAccess); ok {
		t.Fatal("deleted plugin access schema remains")
	}
	if result := deployFixture(t, oldFactory, server.URL+"/deploy", results); result.Version != "1.0.0" {
		t.Fatalf("previously acquired factory lost its deployment version: %+v", result)
	}
}

func TestE2E_MarketMissingReleaseFailsWithoutStoppingService(t *testing.T) {
	const providerType = "e2e-market-no-release"
	var indexMu sync.RWMutex
	index := `{"plugins":[{"provider_type":"e2e-market-no-release","version":"1.0.0","binary":"webhook-deployer","access_provider_type":"e2e-market-no-release-access","protocol_version":1}]}`
	release := buildMarketRelease(t, providerType, "1.0.0", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/index.json" {
			indexMu.RLock()
			defer indexMu.RUnlock()
			_, _ = io.WriteString(w, index)
			return
		}
		if strings.Contains(req.URL.Path, "/releases/download/") {
			_, _ = w.Write(release.archive)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	root := t.TempDir()
	catalog := pluginhost.NewCatalog()
	t.Cleanup(catalog.Close)
	previous := pluginhost.GlobalReloader()
	pluginhost.SetGlobalReloader(pluginhost.NewReloader(plugin.PluginConfig{PluginDir: root}, catalog, slog.Default()))
	t.Cleanup(func() { pluginhost.SetGlobalReloader(previous) })
	market := pluginhost.NewMarketService(pluginhost.MarketConfig{PluginDir: root, IndexURL: server.URL + "/index.json", DownloadMirror: server.URL, CacheTTL: time.Nanosecond})
	job, err := market.Install(context.Background(), providerType)
	if err != nil {
		t.Fatal(err)
	}
	if status := marketJobResult(t, job); status.State != pluginhost.JobFailed || !strings.Contains(status.Error, "release") {
		t.Fatalf("missing release not reported as a failed job: %+v", status)
	}
	if len(catalog.Entries()) != 0 {
		t.Fatal("invalid release appeared in catalog")
	}
	valid, err := json.Marshal(map[string]any{"plugins": []plugin.MarketManifest{release.manifest}})
	if err != nil {
		t.Fatal(err)
	}
	indexMu.Lock()
	index = string(valid)
	indexMu.Unlock()
	retry, err := market.Install(context.Background(), providerType)
	if err != nil {
		t.Fatal(err)
	}
	if status := marketJobResult(t, retry); status.State != pluginhost.JobInstalled {
		t.Fatalf("service did not recover for valid install: %+v", status)
	}
}

func TestE2E_MarketCannotReplaceManualPluginInAnotherDirectory(t *testing.T) {
	const providerType = "e2e-manual-conflict"
	root := t.TempDir()
	dir := setupE2EPlugin(t, root, providerType, providerType+"-access", map[string]any{"schemaRevision": "manual"})
	manualDir := filepath.Join(root, "custom-folder")
	if err := os.Rename(dir, manualDir); err != nil {
		t.Fatal(err)
	}
	cfg := plugin.PluginConfig{PluginDir: root}
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	t.Cleanup(catalog.Close)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	previous := pluginhost.GlobalReloader()
	pluginhost.SetGlobalReloader(pluginhost.NewReloader(cfg, catalog, slog.Default()))
	t.Cleanup(func() { pluginhost.SetGlobalReloader(previous) })
	release := buildMarketRelease(t, providerType, "2.0.0", "")
	results := make(chan webhookResult, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/index.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"plugins": []plugin.MarketManifest{release.manifest}})
		case "/deploy":
			var result webhookResult
			if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
				t.Error(err)
			}
			results <- result
		default:
			_, _ = w.Write(release.archive)
		}
	}))
	defer server.Close()
	market := pluginhost.NewMarketService(pluginhost.MarketConfig{PluginDir: root, IndexURL: server.URL + "/index.json", DownloadMirror: server.URL})
	job, err := market.Install(context.Background(), providerType)
	if err != nil {
		t.Fatal(err)
	}
	if status := marketJobResult(t, job); status.State != pluginhost.JobFailed {
		t.Fatalf("market replaced manual plugin: %+v", status)
	}
	if _, err := os.Stat(filepath.Join(root, providerType)); !os.IsNotExist(err) {
		t.Fatalf("market installed a duplicate provider directory: %v", err)
	}
	factory, err := deployers.Registries.Get(domain.DeploymentProviderType(providerType))
	if err != nil {
		t.Fatal(err)
	}
	if result := deployFixture(t, factory, server.URL+"/deploy", results); result.Version != "manual" {
		t.Fatalf("manual plugin replaced: %+v", result)
	}
}
