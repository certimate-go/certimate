package pluginhost_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/pluginhost/internal/markettest"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/routes"
	_ "github.com/certimate-go/certimate/migrations"
	"github.com/certimate-go/certimate/pkg/plugin"
	"github.com/certimate-go/certimate/ui"
)

const (
	fixtureEmail          = "markets@example.test"
	fixturePassword       = "market-fixture-password123"
	fixtureProvider       = "market-fixture"
	fixtureAccess         = "market-fixture-access"
	fixtureRepoA          = "fixture-a/plugins"
	fixtureRepoB          = "fixture-b/plugins"
	fixtureConsumer       = "market-consumer"
	fixtureDuplicateOwner = "market-duplicate-owner"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type fixtureRelease struct {
	index    http.Handler
	releases []marketReleaseFixture
}

func makeFixtureRelease(repo string, releases ...marketReleaseFixture) fixtureRelease {
	for i := range releases {
		copyOfRelease := *releases[i].manifest.Release
		copyOfRelease.Repo = repo
		releases[i].manifest.Release = &copyOfRelease
	}
	index := markettest.WithReleaseDetails(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := make([]plugin.MarketManifest, 0, len(releases))
		for _, release := range releases {
			entries = append(entries, release.manifest)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"plugins": entries})
	}))
	index.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/index.json", nil))
	return fixtureRelease{index: index, releases: releases}
}

func TestMarketBrowserFixture(t *testing.T) {
	address := os.Getenv("CERTIMATE_TEST_MARKET_ADDRESS")
	if address == "" {
		t.Skip("standalone fixture; set CERTIMATE_TEST_MARKET_ADDRESS")
	}
	state := os.Getenv("CERTIMATE_TEST_MARKET_STATE")
	if state == "" {
		t.Fatal("CERTIMATE_TEST_MARKET_STATE must be a dedicated fixture directory")
	}
	if err := os.MkdirAll(state, 0o700); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{os.Args[0], "--dir", filepath.Join(state, "db")}
	pb := app.GetApp()
	if err := pb.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	defer pb.ResetBootstrapState()
	if err := pb.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	adminCollection, err := pb.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pb.FindAuthRecordByEmail(adminCollection, fixtureEmail); err != nil {
		admin := core.NewRecord(adminCollection)
		admin.SetEmail(fixtureEmail)
		admin.SetPassword(fixturePassword)
		if err := pb.Save(admin); err != nil {
			t.Fatal(err)
		}
	}
	users, err := pb.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pb.FindAuthRecordByEmail(users, "regular@example.test"); err != nil {
		user := core.NewRecord(users)
		user.SetEmail("regular@example.test")
		user.SetPassword(fixturePassword)
		if err := pb.Save(user); err != nil {
			t.Fatal(err)
		}
	}
	root := filepath.Join(state, "plugins")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(state, "seeded")); os.IsNotExist(err) {
		for _, name := range []string{"fixture-legacy", "fixture-manual"} {
			setupE2EPlugin(t, root, name, name+"-access", map[string]any{"schemaRevision": "0.1.0", "accessField": "token", "deployField": "message"})
		}
		if err := plugin.WriteMarketMeta(root, "fixture-legacy", plugin.NewMarketMeta("official", "0.1.0", "0.1.0", "v0.1.0")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(state, "seeded"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cfg := plugin.PluginConfig{PluginDir: root, CoreVersion: app.AppVersion}
	catalog, failures := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	if len(failures) != 0 {
		t.Fatal(failures)
	}
	pluginhost.SetGlobalCatalog(catalog)
	defer catalog.Close()
	reloader := pluginhost.NewReloader(cfg, catalog, slog.Default())
	pluginhost.SetGlobalReloader(reloader)
	defer reloader.Close()

	notifyRelease := func(provider, access, version, failure string, reuse bool) marketReleaseFixture {
		return buildMarketReleaseCapability(t, provider, access, version, failure, reuse, plugin.CapabilityNotify)
	}
	a1 := makeFixtureRelease("fixture-a/releases", notifyRelease(fixtureProvider, fixtureAccess, "1.0.0", "", false))
	a2 := makeFixtureRelease("fixture-a/releases", notifyRelease(fixtureProvider, fixtureAccess, "2.0.0", "", false))
	broken := makeFixtureRelease("fixture-a/releases", notifyRelease(fixtureProvider, fixtureAccess, "2.0.0", "broken fixture", false))
	b := makeFixtureRelease("fixture-b/releases", notifyRelease(fixtureProvider, fixtureAccess, "9.0.0", "", false), notifyRelease(fixtureConsumer, fixtureAccess, "1.0.0", "", true), notifyRelease(fixtureDuplicateOwner, fixtureAccess, "1.0.0", "", false))
	official := makeFixtureRelease("certimate-go/plugins", buildMarketRelease(t, "fixture-legacy", "1.0.0", ""), buildMarketRelease(t, "fixture-manual", "1.0.0", ""))
	var mu sync.RWMutex
	currentA := a1
	invalidA := false
	delay := 2 * time.Second
	markets := map[string]fixtureRelease{fixtureRepoB: b, "certimate-go/plugins": official}
	independent := os.Getenv("CERTIMATE_TEST_INDEPENDENT_CATALOG")
	var independentCatalog, independentReleases http.Handler
	if independent != "" {
		independentCatalog = http.FileServer(http.Dir(filepath.Join(independent, "market")))
		independentReleases = http.FileServer(http.Dir(filepath.Join(independent, "assets")))
	}
	fixtureContent := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 3)
		if len(parts) != 3 {
			http.NotFound(w, r)
			return
		}
		repo := parts[0] + "/" + parts[1]
		mu.RLock()
		selected, invalid := currentA, invalidA
		mu.RUnlock()
		if repo == "independent-fixture/catalog" && independentCatalog != nil && r.URL.Host == "raw.githubusercontent.com" {
			http.StripPrefix("/independent-fixture/catalog/refs/heads/main", independentCatalog).ServeHTTP(w, r)
			return
		}
		if repo == "independent-fixture/releases" && independentReleases != nil && r.URL.Host == "github.com" {
			asset := filepath.Base(r.URL.Path)
			http.StripPrefix(strings.TrimSuffix(r.URL.Path, "/"+asset), independentReleases).ServeHTTP(w, r)
			return
		}
		if r.URL.Host == "raw.githubusercontent.com" {
			if repo != fixtureRepoA {
				selected = markets[repo]
				invalid = false
			}
			if selected.index == nil {
				http.NotFound(w, r)
				return
			}
			if invalid {
				_, _ = w.Write([]byte("invalid catalog"))
				return
			}
			local := r.Clone(r.Context())
			local.URL.Path = strings.TrimPrefix(parts[2], "refs/heads/main")
			selected.index.ServeHTTP(w, local)
			return
		}
		if r.Header.Get("Range") != "" {
			http.Error(w, "fixture asset does not support ranges", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		switch repo {
		case "fixture-b/releases":
			selected = b
		case "certimate-go/plugins":
			selected = official
		case "fixture-a/releases":
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		default:
			http.NotFound(w, r)
			return
		}
		for _, item := range selected.releases {
			if r.URL.Path == "/"+repo+"/releases/download/"+item.manifest.Release.Tag+"/plugin.zip" {
				_, _ = w.Write(item.archive)
				return
			}
		}
		http.NotFound(w, r)
	})
	transport := fixtureTransport(func(req *http.Request) (*http.Response, error) {
		writer := httptest.NewRecorder()
		fixtureContent.ServeHTTP(writer, req)
		return writer.Result(), nil
	})
	svc := pluginhost.NewMarketService(pluginhost.MarketConfig{
		Sources: repository.NewPluginMarketSources(pb), Cache: repository.NewPluginMarketCache(pb),
		PluginDir: root, CatalogTransport: transport, DownloadTransport: transport, CacheTTL: time.Hour,
	})
	pluginhost.SetGlobalMarketService(svc)
	router, err := apis.NewRouter(pb)
	if err != nil {
		t.Fatal(err)
	}
	routes.BindRouter(router)
	router.GET("/{path...}", apis.Static(ui.DistDirFS, false))
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	outer := http.NewServeMux()
	outer.Handle("/", mux)
	outer.HandleFunc("POST /fixture/catalog/{state}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.PathValue("state") {
		case "valid":
			currentA, invalidA = a2, false
		case "broken":
			currentA, invalidA = broken, false
		case "invalid":
			invalidA = true
		default:
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	t.Logf("MARKET_FIXTURE_READY http://%s", listener.Addr())
	server := &http.Server{Handler: outer, ReadHeaderTimeout: 5 * time.Second}
	if err := server.Serve(listener); err != nil {
		t.Fatal(fmt.Errorf("fixture server: %w", err))
	}
}
