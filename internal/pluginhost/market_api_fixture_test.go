package pluginhost_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
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
	"github.com/certimate-go/certimate/internal/setup"
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
	fixturePurger         = "market-purger"
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
		entries := make([]map[string]any, 0, len(releases))
		for _, release := range releases {
			body, _ := json.Marshal(release.manifest)
			var entry map[string]any
			_ = json.Unmarshal(body, &entry)
			if release.manifest.ProviderType == fixtureProvider {
				name := "甲市场通知插件"
				if repo == "fixture-b/releases" {
					name = "乙市场通知插件"
				}
				entry["description"] = "Legacy notification description"
				entry["i18n"] = map[string]map[string]string{
					"zh":    {"name": name, "description": "发送通知到目标服务"},
					"zh-CN": {"name": name + "简体"},
					"en":    {"name": "Market notification plugin", "description": "Send notifications to the target service"},
				}
			}
			entries = append(entries, entry)
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
	b := makeFixtureRelease("fixture-b/releases", notifyRelease(fixtureProvider, fixtureAccess, "9.0.0", "", false), notifyRelease(fixtureConsumer, fixtureAccess, "1.0.0", "", true), notifyRelease(fixtureDuplicateOwner, fixtureAccess, "1.0.0", "", false), buildMarketReleaseCapability(t, fixturePurger, fixtureAccess, "1.0.0", "", true, plugin.CapabilityPurge))
	now := time.Now().UTC()
	legacy := buildMarketRelease(t, "fixture-legacy", "1.0.0", "")
	picture := image.NewRGBA(image.Rect(0, 0, 1, 1))
	picture.Set(0, 0, color.RGBA{B: 240, A: 255})
	var icon bytes.Buffer
	if err := png.Encode(&icon, picture); err != nil {
		t.Fatal(err)
	}
	legacy.marketIcon = icon.Bytes()
	official := makeOfficialCDNFixture(t, 1, uint64(now.UnixMicro()), now, now.Add(24*time.Hour),
		legacy, buildMarketRelease(t, "fixture-manual", "1.0.0", ""))
	officialRawFiles := make(map[string][]byte, len(official.files))
	for path, body := range official.files {
		repo, refPath, _ := strings.Cut(strings.TrimPrefix(path, "/gh/"), "@")
		ref, relative, _ := strings.Cut(refPath, "/")
		if ref == "main" {
			ref = "refs/heads/main"
		}
		officialRawFiles["/"+repo+"/"+ref+"/"+relative] = body
	}
	var signedArchive http.Handler
	if directory := os.Getenv("CERTIMATE_TEST_OFFICIAL_ARCHIVE"); directory != "" {
		root, err := os.OpenRoot(directory)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()
		trust, err := root.ReadFile("trust.json")
		if err != nil {
			t.Fatal(err)
		}
		official.trust = plugin.OfficialMarketTrust{}
		if err := json.Unmarshal(trust, &official.trust); err != nil {
			t.Fatal(err)
		}
		signedArchive = http.StripPrefix("/mirror", http.FileServerFS(root.FS()))
	}
	var mu sync.RWMutex
	currentA := a1
	invalidA := false
	delay := 2 * time.Second
	markets := map[string]fixtureRelease{fixtureRepoB: b}
	independent := os.Getenv("CERTIMATE_TEST_INDEPENDENT_CATALOG")
	var independentCatalog, independentReleases http.Handler
	if independent != "" {
		independentCatalog = http.FileServer(http.Dir(filepath.Join(independent, "market")))
		independentReleases = http.FileServer(http.Dir(filepath.Join(independent, "assets")))
	}
	fixtureContent := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if signedArchive != nil {
			if r.URL.Host == "custom.fixture.example" && strings.HasPrefix(r.URL.Path, "/mirror/gh/") {
				signedArchive.ServeHTTP(w, r)
			} else {
				http.Error(w, "archive fixture only serves the selected mirror", http.StatusBadGateway)
			}
			return
		}
		switch r.URL.Host {
		case "raw.githubusercontent.com":
			if body, ok := officialRawFiles[r.URL.Path]; ok {
				_, _ = w.Write(body)
				return
			}
		case "cdn.jsdelivr.net", "preset.fixture.example", "custom.fixture.example", "tampered.fixture.example":
			prefix := ""
			if r.URL.Host != "cdn.jsdelivr.net" {
				prefix = "/mirror"
			}
			path, _ := strings.CutPrefix(r.URL.Path, prefix)
			if body, ok := official.files[path]; ok && strings.HasPrefix(r.URL.Path, prefix+"/gh/") {
				if r.URL.Host == "tampered.fixture.example" && strings.Contains(path, "/market/icons/") {
					body = append(bytes.Clone(body), 0)
				}
				_, _ = w.Write(body)
			} else {
				http.NotFound(w, r)
			}
			return
		case "offline.fixture.example":
			w.WriteHeader(http.StatusBadGateway)
			return
		}
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
		Settings: repository.NewSettingsRepository(),
		MirrorPresets: []pluginhost.MarketMirrorPreset{
			{ID: "fixture", Name: "Fixture CDN", BaseURL: "https://preset.fixture.example/mirror"},
		},
		PluginDir: root, CatalogTransport: transport, DownloadTransport: transport, CacheTTL: time.Hour, OfficialTrust: official.trust,
	})
	pluginhost.SetGlobalMarketService(svc)
	router, err := apis.NewRouter(pb)
	if err != nil {
		t.Fatal(err)
	}
	routes.BindRouter(router, setup.NewService(pb))
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
