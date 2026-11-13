package repository

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"

	_ "github.com/certimate-go/certimate/migrations"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestPluginMarketSourcesUpgradeAndPersistence(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	open := func() *core.BaseApp {
		t.Helper()
		app := core.NewBaseApp(core.BaseAppConfig{DataDir: dir})
		if err := app.Bootstrap(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = app.ResetBootstrapState() })
		return app
	}
	ctx := context.Background()
	app := open()
	if err := app.RunSystemMigrations(); err != nil {
		t.Fatal(err)
	}
	var earlier core.MigrationsList
	for _, migration := range core.AppMigrations.Items() {
		if migration.File < "1794555000" {
			earlier.Add(migration)
		}
	}
	if _, err := core.NewMigrationsRunner(app, earlier).Up(); err != nil {
		t.Fatal(err)
	}
	cache := NewPluginMarketCache(app)
	official := plugin.OfficialMarketSource().IndexURL()
	syncInfo := PluginMarketSync{SyncedAt: time.Now(), ETag: "pre-upgrade"}
	if err := cache.Save(ctx, official, []*plugin.MarketManifest{{Manifest: plugin.Manifest{ProviderType: "cached"}}}, syncInfo); err != nil {
		t.Fatal(err)
	}
	settings, err := app.FindCollectionByNameOrId("settings")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(settings)
	record.Set("name", "fixture-setting")
	record.Set("content", map[string]string{"value": "existing"})
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}

	repo := NewPluginMarketSources(app)
	if sources, err := repo.List(ctx); err != nil || len(sources) != 0 {
		t.Fatalf("migration seeded sources: %v %v", sources, err)
	}
	cached, metadata, err := cache.Load(ctx, official)
	if err != nil || len(cached) != 1 || cached[0].ProviderType != "cached" || metadata.ETag != syncInfo.ETag {
		t.Fatalf("cache changed: %v %+v %v", cached, metadata, err)
	}
	if saved, err := app.FindRecordById(settings.Id, record.Id); err != nil || saved.GetString("name") != "fixture-setting" {
		t.Fatalf("settings changed: %v", err)
	}

	source, _ := plugin.ParseMarketSource("https://github.com/Example/Plugins")
	alias, _ := plugin.ParseMarketSource("https://github.com/example/plugins.git/")
	outcomes := make(chan error, 2)
	var additions sync.WaitGroup
	for _, candidate := range []plugin.MarketSource{source, alias} {
		additions.Go(func() { outcomes <- repo.Add(ctx, candidate) })
	}
	additions.Wait()
	close(outcomes)
	var successes int
	for err := range outcomes {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent adds succeeded %d times", successes)
	}
	if err := app.ResetBootstrapState(); err != nil {
		t.Fatal(err)
	}

	repo = NewPluginMarketSources(open())
	sources, err := repo.List(ctx)
	if err != nil || len(sources) != 1 || sources[0] != source {
		t.Fatalf("restart sources: %v %v", sources, err)
	}
}
