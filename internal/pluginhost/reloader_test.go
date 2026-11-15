package pluginhost_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestReloader_PurgeLifecycle(t *testing.T) {
	vendor := &purgeVendorService{count: 2}
	srv := newPurgeVendorServer(t, vendor)

	root := t.TempDir()
	fixture := map[string]any{
		"capability": plugin.CapabilityPurge, "schemaRevision": "v1", "omitAccessSchema": true,
		"providerType": purgePluginProvider, "accessProviderType": purgeAccessType,
	}
	dir := setupE2EPlugin(t, root, purgePluginProvider, purgeAccessType, fixture)
	cfg := plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	t.Cleanup(catalog.Close)
	reloader := pluginhost.NewReloader(cfg, catalog, slog.Default())

	res, err := purgeViaClient(t, purgePluginProvider, srv.URL, 24*time.Hour)
	if err != nil || res.Count != 2 {
		t.Fatalf("initial purge: %+v %v", res, err)
	}

	fixture["bootstrapError"] = "purger update is broken"
	writeFixtureJSON(t, filepath.Join(dir, "fixture.json"), fixture)
	failed := reloader.ReloadNow(context.Background())
	if len(failed.Errors) == 0 {
		t.Fatal("broken purge update must fail")
	}
	if _, err := purgers.Registries.Get(domain.PurgeProviderType(purgePluginProvider)); err != nil {
		t.Fatalf("failed update removed the working purge factory: %v", err)
	}
	res, err = purgeViaClient(t, purgePluginProvider, srv.URL, 24*time.Hour)
	if err != nil || res.Count != 2 {
		t.Fatalf("purge after failed update must keep working: %+v %v", res, err)
	}

	delete(fixture, "bootstrapError")
	fixture["schemaRevision"] = "v2"
	writeFixtureJSON(t, filepath.Join(dir, "fixture.json"), fixture)
	updated := reloader.ReloadNow(context.Background())
	if len(updated.Errors) != 0 || len(updated.Changed) != 1 {
		t.Fatalf("purge update failed: %+v", updated)
	}
	if _, err := purgeViaClient(t, purgePluginProvider, srv.URL, 24*time.Hour); err != nil {
		t.Fatalf("purge after update: %v", err)
	}
	calls := vendor.snapshot()
	if len(calls) != 3 {
		t.Fatalf("vendor saw %d calls, want 3", len(calls))
	}
	if calls[2].Version != "v2" {
		t.Fatalf("new calls must use the updated factory, got version %q", calls[2].Version)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	removed := reloader.ReloadNow(context.Background())
	if len(removed.Errors) != 0 || len(removed.Removed) != 1 {
		t.Fatalf("purge removal failed: %+v", removed)
	}
	if _, err := purgers.Registries.Get(domain.PurgeProviderType(purgePluginProvider)); err == nil {
		t.Fatal("removed purge plugin remains registered")
	}
	if _, err := purgeViaClient(t, purgeLegacyProvider, srv.URL, 0); err == nil {
		t.Fatal("legacy id must not purge after plugin removal")
	}
}
