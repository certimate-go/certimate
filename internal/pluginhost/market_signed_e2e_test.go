package pluginhost_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type officialCDNFixture struct {
	trust plugin.OfficialMarketTrust
	files map[string][]byte
}

func makeOfficialCDNFixture(t *testing.T, epoch uint32, generation uint64, issuedAt, expiresAt time.Time, releases ...marketReleaseFixture) officialCDNFixture {
	t.Helper()
	seed := bytes.Repeat([]byte{0x51}, ed25519.SeedSize)
	key := ed25519.NewKeyFromSeed(seed)
	const keyID = "fixture-ed25519"
	if epoch > 1 {
		versionedSeed := sha256.Sum256([]byte(fmt.Sprintf("fixture epoch %d", epoch)))
		key = ed25519.NewKeyFromSeed(versionedSeed[:])
	}
	trust := plugin.OfficialMarketTrust{Epoch: epoch, Keys: map[string]ed25519.PublicKey{keyID: key.Public().(ed25519.PublicKey)}}
	files := make(map[string][]byte)
	summaries := make([]map[string]any, 0, len(releases))
	platforms := []string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm", "linux/arm64", "windows/386", "windows/amd64", "windows/arm64"}
	indexDir := strings.TrimSuffix(plugin.OfficialMarketIndexPath, "index.json")
	for _, release := range releases {
		manifest := release.manifest.Manifest
		repo := "certimate-go/plugin-provider-" + manifest.ProviderType + "-fixture"
		artifactCommit := fmt.Sprintf("%x", sha256.Sum256(release.archive))[:40]
		digest := fmt.Sprintf("%x", sha256.Sum256(release.archive))
		artifacts := make([]plugin.OfficialMarketArtifact, 0, len(platforms))
		for _, platform := range platforms {
			platformOS, arch, _ := strings.Cut(platform, "/")
			path := "build/" + platformOS + "_" + arch + ".zip"
			artifacts = append(artifacts, plugin.OfficialMarketArtifact{
				OS: platformOS, Arch: arch, Path: path, Size: int64(len(release.archive)), SHA256: digest,
			})
			if platformOS == runtime.GOOS && arch == runtime.GOARCH {
				files["/gh/"+repo+"@"+artifactCommit+"/"+path] = release.archive
			}
		}
		detailFields := map[string]any{
			"provider_type": manifest.ProviderType, "repository": repo, "version": manifest.Version, "tag": "v" + manifest.Version,
			"source_commit": strings.Repeat("e", 40), "release_commit": artifactCommit, "build_run_id": "17",
			"manifest": manifest, "artifacts": artifacts,
		}
		var icon *plugin.OfficialMarketIcon
		if len(release.marketIcon) != 0 {
			icon = &plugin.OfficialMarketIcon{Format: "png", SHA256: fmt.Sprintf("%x", sha256.Sum256(release.marketIcon))}
			detailFields["icon"] = icon
			files[indexDir+"market/icons/"+icon.SHA256+".png"] = release.marketIcon
		}
		detail, err := json.Marshal(detailFields)
		if err != nil {
			t.Fatal(err)
		}
		detailSHA := fmt.Sprintf("%x", sha256.Sum256(detail))
		files[indexDir+"market/details/"+detailSHA+".json"] = detail
		summaries = append(summaries, map[string]any{
			"provider_type": manifest.ProviderType, "version": manifest.Version, "capability": manifest.Capability,
			"category": manifest.Category, "display_name_key": manifest.DisplayNameKey, "description": manifest.Description,
			"protocol_version": manifest.ProtocolVersion, "repository": repo, "release_commit": artifactCommit,
			"platforms": platforms, "detail": map[string]string{"sha256": detailSHA},
			"icon": icon,
		})
	}
	payload, err := json.Marshal(map[string]any{
		"format_version": 1, "epoch": epoch, "generation": generation, "issued_at": issuedAt, "expires_at": expiresAt,
		"source_commit": strings.Repeat("1", 40), "plugins": summaries,
	})
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := json.Marshal(map[string]any{
		"key_id": keyID, "payload": payload, "signature": ed25519.Sign(key, append([]byte("certimate:plugin-market:index:v1\x00"), payload...)),
	})
	if err != nil {
		t.Fatal(err)
	}
	files[plugin.OfficialMarketIndexPath] = envelope
	return officialCDNFixture{trust: trust, files: files}
}

func TestE2E_SignedOfficialMarketInstallAndRejectTampering(t *testing.T) {
	const provider, alias = "e2e-signed-market", "e2e-signed-market-legacy"
	now := time.Now().UTC().Truncate(time.Second)
	first := buildMarketReleaseCapability(t, provider, provider+"-access", "1.0.0", "", false, plugin.CapabilityDns, alias)
	first.manifest.ProviderAliases = nil
	next := buildMarketReleaseCapability(t, provider, provider+"-access", "2.0.0", "", false, plugin.CapabilityDns)
	signed := makeOfficialCDNFixture(t, 1, 11, now, now.Add(time.Hour), first)
	update := makeOfficialCDNFixture(t, 1, 12, now, now.Add(time.Hour), next)
	var archivePath string
	for path := range signed.files {
		if strings.HasSuffix(path, ".zip") {
			archivePath = path
			break
		}
	}
	if archivePath == "" {
		t.Fatal("fixture is missing the official archive")
	}
	rawArchive := plugin.GitHubRawBase + "/" + strings.Replace(strings.TrimPrefix(archivePath, "/gh/"), "@", "/", 1)
	var mu sync.Mutex
	files := signed.files
	mode := "unsigned"
	transport := fixtureTransport(func(request *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()
		writer := httptest.NewRecorder()
		path := request.URL.Path
		if request.URL.String() == rawArchive && mode == "redirect" {
			path = archivePath
		} else if request.URL.Host != "cdn.jsdelivr.net" {
			t.Errorf("official market asked another host: %s", request.URL)
		} else if strings.HasSuffix(path, ".zip") && (mode == "redirect" || mode == "bad-redirect") {
			destination := plugin.GitHubRawBase + "/" + strings.Replace(strings.TrimPrefix(path, "/gh/"), "@", "/", 1)
			if mode == "bad-redirect" {
				destination = strings.Replace(rawArchive, "/build/", "/other/", 1)
			}
			writer.Header().Set("Location", destination)
			writer.WriteHeader(http.StatusMovedPermanently)
			return writer.Result(), nil
		}
		if path == plugin.OfficialMarketIndexPath && mode == "unsigned" {
			fmt.Fprintln(writer, `{"plugins":[]}`)
		} else if body, ok := files[path]; ok && !(mode == "missing-detail" && strings.Contains(path, "/details/")) {
			if mode == "bad-detail" && strings.Contains(path, "/details/") || mode == "bad-zip" && strings.HasSuffix(path, ".zip") {
				body = append(bytes.Clone(body[:len(body)-1]), body[len(body)-1]^1)
			}
			writer.Write(body)
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
		return writer.Result(), nil
	})
	root := t.TempDir()
	catalog := pluginhost.NewCatalog()
	t.Cleanup(catalog.Close)
	previousReloader := pluginhost.GlobalReloader()
	pluginhost.SetGlobalReloader(pluginhost.NewReloader(plugin.PluginConfig{PluginDir: root}, catalog, slog.Default()))
	t.Cleanup(func() { pluginhost.SetGlobalReloader(previousReloader) })
	market := pluginhost.NewMarketService(pluginhost.MarketConfig{
		PluginDir: root, OfficialTrust: signed.trust, CatalogTransport: transport, DownloadTransport: transport,
		DownloadMirror: "https://third-party-only.test",
	})
	setMode := func(value string) {
		mu.Lock()
		defer mu.Unlock()
		mode = value
	}
	install := func(state pluginhost.JobState, contains string) {
		t.Helper()
		job, err := market.Install(t.Context(), provider)
		if err != nil {
			t.Fatal(err)
		}
		status := marketJobResult(t, job)
		if status.State != state || !strings.Contains(status.Error, contains) {
			t.Fatalf("install gave %+v; want %s and %q", status, state, contains)
		}
	}
	install(pluginhost.JobFailed, "signed index")
	setMode("missing-detail")
	if status, err := market.RefreshSource(t.Context(), "official"); err != nil || status.Error != "" {
		t.Fatalf("refresh signed index: %+v, %v", status, err)
	}
	install(pluginhost.JobFailed, "status 404")
	setMode("bad-detail")
	install(pluginhost.JobFailed, "detail checksum mismatch")
	setMode("bad-zip")
	install(pluginhost.JobFailed, "checksum mismatch")
	setMode("redirect")
	install(pluginhost.JobInstalled, "")
	assertVersion := func(want string) {
		t.Helper()
		meta, err := plugin.ReadMarketMeta(root, provider)
		if err != nil || meta.InstalledVersion != want || meta.Origin() != plugin.OfficialMarketSource() {
			t.Fatalf("installed metadata: %+v, %v; want version %s", meta, err, want)
		}
	}
	assertVersion("1.0.0")
	if dp := catalog.DNSPlugin(alias); dp == nil || dp.Manifest.ProviderType != provider {
		t.Fatal("installed signed archive did not register its DNS alias")
	}
	mu.Lock()
	files, mode = update.files, "bad-redirect"
	mu.Unlock()
	if status, err := market.RefreshSource(t.Context(), "official"); err != nil || status.Error != "" {
		t.Fatalf("refresh update: %+v, %v", status, err)
	}
	if _, err := market.Update(t.Context(), provider, "official"); err == nil || !strings.Contains(err.Error(), "unexpected catalog redirect") {
		t.Fatalf("unexpected update redirect result: %v", err)
	}
	setMode("bad-zip")
	if _, err := market.Update(t.Context(), provider, "official"); !errors.Is(err, pluginhost.ErrOfficialChecksumMismatch) {
		t.Fatalf("tampered update: %v", err)
	}
	assertVersion("1.0.0")
	setMode("")
	if _, err := market.Update(t.Context(), provider, "official"); err != nil {
		t.Fatal(err)
	}
	assertVersion("2.0.0")
	if dp := catalog.DNSPlugin(alias); dp != nil {
		t.Fatal("update left behind a removed DNS alias")
	}
}

func TestE2E_OfficialIndexHighWaterSurvivesMirrorAndRestart(t *testing.T) {
	directory := t.TempDir()
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: directory})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	now := time.Now().UTC().Truncate(time.Second)
	high := makeOfficialCDNFixture(t, 1, 10, now, now.Add(time.Hour))
	low := makeOfficialCDNFixture(t, 1, 9, now, now.Add(time.Hour))
	conflict := makeOfficialCDNFixture(t, 1, 10, now, now.Add(2*time.Hour))
	newEpoch := makeOfficialCDNFixture(t, 2, 1, now, now.Add(time.Hour))
	expired := makeOfficialCDNFixture(t, 2, 2, now.Add(-2*time.Hour), now.Add(-time.Hour))
	index := func(mirror string, cached bool, fixture officialCDNFixture, code int) *pluginhost.MarketService {
		return pluginhost.NewMarketService(pluginhost.MarketConfig{
			PluginDir: t.TempDir(), Cache: repository.NewPluginMarketCache(app),
			IndexURL: mirror + plugin.OfficialMarketIndexPath, OfficialTrust: fixture.trust,
			CatalogTransport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
				if conditional := request.Header.Get("If-None-Match"); cached != (conditional == "fixture-etag") {
					t.Errorf("conditional validator for %s: %q, expected sent=%t", mirror, conditional, cached)
				}
				writer := httptest.NewRecorder()
				writer.Header().Set("ETag", "fixture-etag")
				writer.WriteHeader(code)
				if code == http.StatusOK {
					writer.Write(fixture.files[plugin.OfficialMarketIndexPath])
				}
				return writer.Result(), nil
			}),
		})
	}
	refresh := func(market *pluginhost.MarketService, message string) {
		t.Helper()
		status, err := market.RefreshSource(t.Context(), "official")
		if err != nil || (message == "" && status.Error != "") || !strings.Contains(status.Error, message) {
			t.Fatalf("wanted %q, got %+v, %v", message, status, err)
		}
	}
	refresh(index("https://first.example", false, high, http.StatusOK), "")
	if err := app.ResetBootstrapState(); err != nil {
		t.Fatal(err)
	}
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	refresh(index("https://first.example", true, high, http.StatusNotModified), "")
	refresh(index("https://second.example", false, low, http.StatusOK), "rollback or conflict")
	refresh(index("https://second.example", false, conflict, http.StatusOK), "rollback or conflict")
	refresh(index("https://second.example", false, newEpoch, http.StatusOK), "")
	refresh(index("https://second.example", true, expired, http.StatusOK), "expired")
	refresh(index("https://first.example", false, high, http.StatusOK), "rollback or conflict")
}

type mirrorSettingsFixture struct {
	fail    bool
	readErr error
	content domain.SettingsContent
}

func (store *mirrorSettingsFixture) GetByName(context.Context, string) (*domain.Settings, error) {
	if store.readErr != nil {
		return nil, store.readErr
	}
	if store.content == nil {
		return nil, domain.ErrRecordNotFound
	}
	return &domain.Settings{Name: domain.SettingsNamePluginMarket, Content: store.content}, nil
}

func (store *mirrorSettingsFixture) Save(_ context.Context, settings *domain.Settings) (*domain.Settings, error) {
	if store.fail {
		return nil, errors.New("fixture database write failed")
	}
	return settings, nil
}

func TestE2E_OfficialMirrorRetiresPendingRefreshAfterPersisting(t *testing.T) {
	now := time.Now().UTC()
	old := makeOfficialCDNFixture(t, 1, 99, now, now.Add(time.Hour))
	current := makeOfficialCDNFixture(t, 1, 2, now, now.Add(time.Hour))
	started, resume := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-resume:
		default:
			close(resume)
		}
	}()
	settings := &mirrorSettingsFixture{fail: true}
	mirror := pluginhost.MarketMirror{Mode: "custom", BaseURL: "https://new-mirror.example"}
	catalog := pluginhost.NewMarketService(pluginhost.MarketConfig{
		Settings: settings, OfficialTrust: old.trust, PluginDir: t.TempDir(),
		CatalogTransport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
			writer := httptest.NewRecorder()
			switch request.URL.Host {
			case "cdn.jsdelivr.net":
				close(started)
				<-resume
				writer.Write(old.files[request.URL.Path])
			case "new-mirror.example":
				writer.Write(current.files[request.URL.Path])
			default:
				writer.WriteHeader(http.StatusNotFound)
			}
			return writer.Result(), nil
		}),
	})
	if _, err := catalog.SetOfficialMirror(t.Context(), mirror); err == nil || catalog.OfficialMirror(t.Context()).Mirror.Mode != "default" {
		t.Fatalf("failed write applied mirror: %v", err)
	}
	pending := make(chan error, 1)
	go func() {
		_, err := catalog.RefreshSource(t.Context(), "official")
		pending <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first mirror was not selected after failed write")
	}
	settings.fail = false
	if _, err := catalog.SetOfficialMirror(t.Context(), mirror); err != nil {
		t.Fatal(err)
	}
	assertCurrent := func() {
		t.Helper()
		if status, err := catalog.RefreshSource(t.Context(), "official"); err != nil || status.Error != "" {
			t.Fatalf("new mirror failed verification: %+v, %v", status, err)
		}
	}
	assertCurrent()
	close(resume)
	if err := <-pending; err != pluginhost.ErrMarketUnavailable {
		t.Fatalf("retired refresh result: %v", err)
	}
	assertCurrent()
}

func TestE2E_OfficialMarketRecoversWhenSettingsReadResumes(t *testing.T) {
	const mirrorHost = "saved-mirror.example"
	const mirrorBase = "https://" + mirrorHost
	failure := errors.New("fixture settings read failed")
	settings := &mirrorSettingsFixture{
		fail: true, readErr: failure,
		content: domain.SettingsContent{"mode": "custom", "baseUrl": mirrorBase},
	}
	now := time.Now().UTC()
	cdn := makeOfficialCDNFixture(t, 1, 2, now, now.Add(time.Hour))
	var requests []string
	market := pluginhost.NewMarketService(pluginhost.MarketConfig{
		Settings: settings, OfficialTrust: cdn.trust, PluginDir: t.TempDir(),
		CatalogTransport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
			requests = append(requests, request.URL.Host)
			writer := httptest.NewRecorder()
			writer.Write(cdn.files[request.URL.Path])
			return writer.Result(), nil
		}),
	})
	if current := market.OfficialMirror(t.Context()); current.Error != failure.Error() {
		t.Fatalf("settings read failure was not reported: %+v", current)
	}
	query := pluginhost.MarketQuery{Source: "official", Refresh: true}
	page, err := market.ListMarketPage(t.Context(), query)
	if err != nil || !page.RefreshFailed || len(requests) != 0 {
		t.Fatalf("settings read failure allowed a mirror request: %+v, %v, %v", page, err, requests)
	}
	settings.readErr = nil
	page, err = market.ListMarketPage(t.Context(), query)
	if err != nil || page.RefreshFailed || len(requests) != 1 || requests[0] != mirrorHost {
		t.Fatalf("saved mirror did not recover without rewriting settings: %+v, %v, %v", page, err, requests)
	}
	if current := market.OfficialMirror(t.Context()); current.Error != "" ||
		current.Mirror != (pluginhost.MarketMirror{Mode: "custom", BaseURL: mirrorBase}) {
		t.Fatalf("saved mirror was not restored: %+v", current)
	}
}

type officialReadCacheFixture struct {
	*repository.PluginMarketCache
	readErr error
}

func (cache *officialReadCacheFixture) Load(ctx context.Context, source string) ([]*plugin.MarketManifest, repository.PluginMarketSync, error) {
	if cache.readErr != nil {
		return nil, repository.PluginMarketSync{}, cache.readErr
	}
	return cache.PluginMarketCache.Load(ctx, source)
}

func TestE2E_OfficialMarketReportsUnavailableVerifiedCache(t *testing.T) {
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	healthy := makeOfficialCDNFixture(t, 1, 2, now, now.Add(time.Hour))
	old := makeOfficialCDNFixture(t, 1, 1, now, now.Add(time.Hour))
	served := healthy
	cache := &officialReadCacheFixture{PluginMarketCache: repository.NewPluginMarketCache(app)}
	requests := 0
	config := pluginhost.MarketConfig{
		Cache: cache, OfficialTrust: healthy.trust, PluginDir: t.TempDir(),
		CatalogTransport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
			requests++
			writer := httptest.NewRecorder()
			writer.Write(served.files[request.URL.Path])
			return writer.Result(), nil
		}),
	}
	status, err := pluginhost.NewMarketService(config).RefreshSource(t.Context(), "official")
	if err != nil || status.Error != "" || requests != 1 {
		t.Fatalf("initial index was not saved: %+v, %v, requests=%d", status, err, requests)
	}
	failure := errors.New("fixture cache read failed")
	cache.readErr = failure
	market := pluginhost.NewMarketService(config)
	query := pluginhost.MarketQuery{Source: "official", Refresh: true}
	page, err := market.ListMarketPage(t.Context(), query)
	if err != nil || !page.RefreshFailed || len(page.Sources) != 1 ||
		!strings.Contains(page.Sources[0].Error, failure.Error()) || requests != 1 {
		t.Fatalf("cache read failure was not reported before network access: %+v, %v, requests=%d", page, err, requests)
	}
	cache.readErr = nil
	served = old
	status, err = market.RefreshSource(t.Context(), "official")
	if err != nil || !strings.Contains(status.Error, "rollback or conflict") || requests != 2 {
		t.Fatalf("restored state did not reject a signed rollback: %+v, %v, requests=%d", status, err, requests)
	}
	served = healthy
	status, err = market.RefreshSource(t.Context(), "official")
	if err != nil || status.Error != "" || requests != 3 {
		t.Fatalf("verified refresh did not recover: %+v, %v, requests=%d", status, err, requests)
	}
}
