package pluginhost

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func newTestService(t *testing.T, handler http.HandlerFunc) *MarketService {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewMarketService(MarketConfig{
		IndexURL:  srv.URL,
		PluginDir: t.TempDir(),
		CacheTTL:  time.Hour,
	})
}

func indexJSON(t *testing.T, plugins ...map[string]any) string {
	t.Helper()
	buf, err := json.Marshal(map[string]any{"plugins": plugins})
	if err != nil {
		t.Fatal(err)
	}
	return string(buf)
}

func entry(providerType, version string, release map[string]any) map[string]any {
	m := map[string]any{"provider_type": providerType, "version": version, "binary": providerType}
	if release != nil {
		m["release"] = release
	}
	return m
}

func releaseWithCurrentAsset() map[string]any {
	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	return map[string]any{
		"repo":      "certimate-go/plugins",
		"tag":       "v1.0.0",
		"assets":    map[string]any{key: "bin-" + key},
		"checksums": map[string]any{key: "deadbeef"},
	}
}

func TestListMarket_HappyPath(t *testing.T) {
	body := indexJSON(t,
		entry("beta", "2.0.0", nil),
		entry("alpha", "1.0.0", releaseWithCurrentAsset()),
	)
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	got, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	if got[0].ProviderType != "beta" || got[0].Version != "2.0.0" {
		t.Fatalf("entry 0 mismatch: %+v", got[0])
	}
	if got[1].ProviderType != "alpha" || got[1].Version != "1.0.0" {
		t.Fatalf("entry 1 mismatch: %+v", got[1])
	}
}

func TestListMarket_Empty(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"plugins":[]}`)
	})

	got, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 entries, got %d", len(got))
	}
}

func TestListMarket_UnsupportedPlatform(t *testing.T) {
	body := indexJSON(t,
		entry("noRelease", "1.0.0", nil),
		entry("wrongAsset", "1.0.0", map[string]any{
			"repo":      "x",
			"tag":       "v1",
			"assets":    map[string]any{"none/none": "bin"},
			"checksums": map[string]any{"none/none": "x"},
		}),
	)
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	got, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	for _, e := range got {
		if e.Status != "unsupported_platform" {
			t.Fatalf("want unsupported_platform for %q, got %q", e.ProviderType, e.Status)
		}
	}
}

func TestListMarket_StatusNotInstalled(t *testing.T) {
	body := indexJSON(t, entry("alpha", "1.0.0", releaseWithCurrentAsset()))
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	got, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Status != "not_installed" {
		t.Fatalf("want one not_installed entry, got %+v", got)
	}
}

func TestListMarket_CacheNoRefetch(t *testing.T) {
	var hits atomic.Int32
	body := indexJSON(t, entry("alpha", "1.0.0", nil))
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		io.WriteString(w, body)
	})

	if _, err := svc.ListMarket(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListMarket(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("want 1 server hit (cached second call), got %d", got)
	}
}

func TestListMarket_StaleCacheOnError(t *testing.T) {
	var fail atomic.Bool
	body := indexJSON(t, entry("alpha", "1.0.0", nil))
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		io.WriteString(w, body)
	})

	first, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	svc.cachedAt = time.Now().Add(-2 * time.Hour)
	fail.Store(true)

	second, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatalf("want stale cache on fetch failure, got error: %v", err)
	}
	if len(second) != len(first) {
		t.Fatalf("stale cache mismatch: first %d, second %d", len(first), len(second))
	}
}

func TestListMarket_404(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := svc.ListMarket(context.Background())
	if err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("want 'not available' error, got %v", err)
	}
}

func TestListMarket_RateLimitedNoCache(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := svc.ListMarket(context.Background())
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("want rate-limit error, got %v", err)
	}
}

func TestListMarket_Malformed(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{not valid json`)
	})

	_, err := svc.ListMarket(context.Background())
	if err == nil {
		t.Fatal("want parse error for malformed index, got nil")
	}
}

func TestGetMarketManifest_FromCache(t *testing.T) {
	body := indexJSON(t, entry("alpha", "1.0.0", releaseWithCurrentAsset()))
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	if _, err := svc.ListMarket(context.Background()); err != nil {
		t.Fatal(err)
	}
	mm, err := svc.GetMarketManifest(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if mm.ProviderType != "alpha" || mm.Release == nil {
		t.Fatalf("unexpected manifest: %+v", mm)
	}
}

type zipEntry struct {
	Name string
	Data []byte
}

func buildZip(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		f, err := w.Create(e.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(e.Data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func alphaManifestJSON(providerType, version string) string {
	return fmt.Sprintf(`{"provider_type":%q,"version":%q,"binary":%q,"protocol_version":%d,"access_provider_type":%q,"deploy_category":"cdn","usages":["hosting"]}`, providerType, version, providerType, plugin.ProtocolVersion, providerType)
}

func alphaIndex(t *testing.T, version, checksum string) string {
	t.Helper()
	key := plugin.AssetKey(runtime.GOOS, runtime.GOARCH)
	asset := "alpha_" + runtime.GOOS + "_" + runtime.GOARCH + ".zip"
	if checksum == "" {
		return fmt.Sprintf(`{"plugins":[{"provider_type":"alpha","version":%q,"binary":"alpha","protocol_version":%d,"release":{"repo":"certimate-go/plugins","tag":"alpha/v%s","assets":{%q:%q}}}]}`, version, plugin.ProtocolVersion, version, key, asset)
	}
	return fmt.Sprintf(`{"plugins":[{"provider_type":"alpha","version":%q,"binary":"alpha","protocol_version":%d,"release":{"repo":"certimate-go/plugins","tag":"alpha/v%s","assets":{%q:%q},"checksums":{%q:%q}}}]}`, version, plugin.ProtocolVersion, version, key, asset, key, checksum)
}

func newZipTestService(t *testing.T, indexBody string, zipBytes []byte) *MarketService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/index.json"):
			io.WriteString(w, indexBody)
		case strings.Contains(r.URL.Path, "/releases/download/"):
			_, _ = w.Write(zipBytes)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return NewMarketService(MarketConfig{
		IndexURL:       srv.URL + "/index.json",
		DownloadMirror: srv.URL,
		PluginDir:      t.TempDir(),
		CacheTTL:       time.Hour,
	})
}

func setupReloader(t *testing.T, pluginDir string) {
	t.Helper()
	cfg := plugin.PluginConfig{PluginDir: pluginDir, StartTimeout: 100 * time.Millisecond}
	r := NewReloader(cfg, NewCatalog(), slog.Default())
	prev := GlobalReloader()
	SetGlobalReloader(r)
	t.Cleanup(func() { SetGlobalReloader(prev) })
}

func waitForJob(t *testing.T, job *InstallJob, timeout time.Duration) JobStatus {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		st := job.Status()
		if st.State.Terminal() {
			return st
		}
		time.Sleep(5 * time.Millisecond)
	}
	return job.Status()
}

func TestInstall_AsyncZip(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("fake binary bytes")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))},
		{Name: "alpha.svg", Data: []byte("<svg/>")},
	})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	setupReloader(t, svc.pluginDir)

	job, err := svc.Install(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	st := waitForJob(t, job, 5*time.Second)
	if st.State != JobInstalled {
		t.Fatalf("want installed, got %s (%s)", st.State, st.Error)
	}

	target := filepath.Join(svc.pluginDir, "alpha")
	if _, err := os.Stat(filepath.Join(target, "alpha")); err != nil {
		t.Fatalf("binary missing after install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "manifest.json")); err != nil {
		t.Fatalf("manifest missing after install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".market.json")); err != nil {
		t.Fatalf("market meta missing after install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(svc.pluginDir, ".tmp-alpha")); err == nil {
		t.Fatal("temp dir was not cleaned up")
	}
}

func TestInstall_ChecksumMismatchFails(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("fake binary")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))},
	})
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", "deadbeef"), zipBytes)
	setupReloader(t, svc.pluginDir)

	job, err := svc.Install(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	st := waitForJob(t, job, 5*time.Second)
	if st.State != JobFailed || !strings.Contains(st.Error, "checksum mismatch") {
		t.Fatalf("want failed checksum mismatch, got %s (%s)", st.State, st.Error)
	}
	if _, err := os.Stat(filepath.Join(svc.pluginDir, "alpha")); err == nil {
		t.Fatal("plugin dir must not exist after failed install")
	}
}

func TestInstall_AbsentChecksumFailsClosed(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("fake binary")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))},
	})
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", ""), zipBytes)
	setupReloader(t, svc.pluginDir)

	job, err := svc.Install(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	st := waitForJob(t, job, 5*time.Second)
	if st.State != JobFailed || !strings.Contains(st.Error, "no published checksum") {
		t.Fatalf("want failed (no published checksum), got %s (%s)", st.State, st.Error)
	}
}

func TestInstall_ZipSlipRejected(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("fake binary")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))},
		{Name: "../evil.txt", Data: []byte("escaped")},
	})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	setupReloader(t, svc.pluginDir)

	job, err := svc.Install(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	st := waitForJob(t, job, 5*time.Second)
	if st.State != JobFailed || !strings.Contains(st.Error, "escapes destination") {
		t.Fatalf("want failed zip-slip, got %s (%s)", st.State, st.Error)
	}
	escaped := filepath.Join(svc.pluginDir, "evil.txt")
	if _, err := os.Stat(escaped); err == nil {
		t.Fatalf("zip-slip wrote outside plugin dir: %s", escaped)
	}
}

func TestInstall_BundledManifestMismatchFails(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("fake binary")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("beta", "1.0.0"))},
	})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	setupReloader(t, svc.pluginDir)

	job, err := svc.Install(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	st := waitForJob(t, job, 5*time.Second)
	if st.State != JobFailed || !strings.Contains(st.Error, "provider_type") {
		t.Fatalf("want failed provider_type mismatch, got %s (%s)", st.State, st.Error)
	}
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{{Name: "alpha", Data: []byte("x")}, {Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))}})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	if err := os.MkdirAll(filepath.Join(svc.pluginDir, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Install(context.Background(), "alpha")
	if err != ErrAlreadyInstalled {
		t.Fatalf("want ErrAlreadyInstalled, got %v", err)
	}
}

func TestInstall_OpInProgressReturns409(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{{Name: "alpha", Data: []byte("x")}, {Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))}})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	if err := svc.ops.claim("alpha", opUpdate); err != nil {
		t.Fatal(err)
	}
	defer svc.ops.release("alpha")

	_, err := svc.Install(context.Background(), "alpha")
	if err == nil {
		t.Fatal("want conflict error, got nil")
	}
	de, ok := err.(*domain.Error)
	if !ok || de.Code != 409 {
		t.Fatalf("want domain 409, got %v", err)
	}
}

func TestDelete_BlockedWhileInstallActive(t *testing.T) {
	zipBytes := buildZip(t, []zipEntry{{Name: "alpha", Data: []byte("x")}, {Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "1.0.0"))}})
	sum := sha256.Sum256(zipBytes)
	svc := newZipTestService(t, alphaIndex(t, "1.0.0", hex.EncodeToString(sum[:])), zipBytes)
	if err := svc.ops.claim("alpha", opInstall); err != nil {
		t.Fatal(err)
	}
	defer svc.ops.release("alpha")

	_, err := svc.Delete(context.Background(), "alpha")
	if err == nil {
		t.Fatal("want conflict error, got nil")
	}
	de, ok := err.(*domain.Error)
	if !ok || de.Code != 409 {
		t.Fatalf("want domain 409, got %v", err)
	}
}

func TestUpdate_UpgradesToNewerVersion(t *testing.T) {
	newZip := buildZip(t, []zipEntry{
		{Name: "alpha", Data: []byte("new fake binary")},
		{Name: "manifest.json", Data: []byte(alphaManifestJSON("alpha", "2.0.0"))},
	})
	sum := sha256.Sum256(newZip)
	svc := newZipTestService(t, alphaIndex(t, "2.0.0", hex.EncodeToString(sum[:])), newZip)
	setupReloader(t, svc.pluginDir)

	installed := filepath.Join(svc.pluginDir, "alpha")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installed, "alpha"), []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installed, "manifest.json"), []byte(alphaManifestJSON("alpha", "1.0.0")), 0o644); err != nil {
		t.Fatal(err)
	}
	oldMeta, _ := json.Marshal(plugin.NewMarketMeta("official", "1.0.0", "1.0.0", "alpha/v1.0.0"))
	if err := os.WriteFile(filepath.Join(installed, ".market.json"), oldMeta, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Update(context.Background(), "alpha"); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	meta, err := plugin.ReadMarketMeta(svc.pluginDir, "alpha")
	if err != nil || meta == nil {
		t.Fatalf("market meta missing after update: %v", err)
	}
	if meta.InstalledVersion != "2.0.0" {
		t.Fatalf("want installed version 2.0.0, got %q", meta.InstalledVersion)
	}
	if _, err := os.Stat(filepath.Join(svc.pluginDir, ".bak-alpha")); err == nil {
		t.Fatal("backup dir was not cleaned up after successful update")
	}
}

func TestSweepOrphans_RemovesTemp(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".tmp-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	SweepOrphans(dir, slog.Default())
	if _, err := os.Stat(filepath.Join(dir, ".tmp-alpha")); err == nil {
		t.Fatal("temp dir should be removed")
	}
}

func TestSweepOrphans_RestoresBackup(t *testing.T) {
	dir := t.TempDir()
	bak := filepath.Join(dir, ".bak-alpha")
	if err := os.MkdirAll(bak, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bak, "manifest.json"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	SweepOrphans(dir, slog.Default())
	if _, err := os.Stat(bak); err == nil {
		t.Fatal("backup dir should be renamed away")
	}
	got, err := os.ReadFile(filepath.Join(dir, "alpha", "manifest.json"))
	if err != nil {
		t.Fatalf("plugin should be restored from backup: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("restored content mismatch: %q", got)
	}
}

func TestSweepOrphans_RemovesStaleBackup(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".bak-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	SweepOrphans(dir, slog.Default())
	if _, err := os.Stat(filepath.Join(dir, ".bak-alpha")); err == nil {
		t.Fatal("stale backup should be removed when target exists")
	}
	if _, err := os.Stat(filepath.Join(dir, "alpha")); err != nil {
		t.Fatal("existing plugin should be untouched")
	}
}

func TestFetchZip_ReportsProgress(t *testing.T) {
	payload := bytes.Repeat([]byte("z"), 150*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	svc := NewMarketService(MarketConfig{IndexURL: srv.URL, PluginDir: t.TempDir()})
	dest := filepath.Join(svc.pluginDir, "test.zip")
	var lastD, lastT int64
	var calls int
	if err := svc.fetchZip(context.Background(), srv.URL, dest, func(downloaded, total int64) {
		lastD = downloaded
		lastT = total
		calls++
	}); err != nil {
		t.Fatalf("fetchZip failed: %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(payload) {
		t.Fatalf("downloaded %d bytes, want %d", len(data), len(payload))
	}
	if calls == 0 {
		t.Fatal("progress callback was never invoked")
	}
	if lastT != int64(len(payload)) {
		t.Fatalf("reported total %d, want %d", lastT, len(payload))
	}
	if lastD != int64(len(payload)) {
		t.Fatalf("final reported downloaded %d, want %d", lastD, len(payload))
	}
}

func TestFetchZip_ResumesAfterTruncation(t *testing.T) {
	payload := bytes.Repeat([]byte("z"), 200*1024)
	cut := int64(100 * 1024)
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		var start int64
		if rng := r.Header.Get("Range"); strings.HasPrefix(rng, "bytes=") {
			fmt.Sscanf(rng, "bytes=%d-", &start)
		}
		if n == 1 && start == 0 {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			_, _ = w.Write(payload[:cut])
			return
		}
		end := int64(len(payload)) - 1
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(payload)))
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(payload))-start, 10))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload[start:])
	}))
	t.Cleanup(srv.Close)

	svc := NewMarketService(MarketConfig{IndexURL: srv.URL, PluginDir: t.TempDir()})
	dest := filepath.Join(svc.pluginDir, "test.zip")
	if err := svc.fetchZip(context.Background(), srv.URL, dest, nil); err != nil {
		t.Fatalf("fetchZip failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("resumed download produced wrong content: got %d bytes, want %d", len(got), len(payload))
	}
	if attempts.Load() < 2 {
		t.Fatalf("expected at least 2 attempts (one truncated, one resumed), got %d", attempts.Load())
	}
}

func TestFetchZip_ResumesAfterStall(t *testing.T) {
	payload := bytes.Repeat([]byte("z"), 200*1024)
	cut := int64(100 * 1024)
	stall := 80 * time.Millisecond

	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		var start int64
		if rng := r.Header.Get("Range"); strings.HasPrefix(rng, "bytes=") {
			fmt.Sscanf(rng, "bytes=%d-", &start)
		}
		if n == 1 && start == 0 {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			_, _ = w.Write(payload[:cut])
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
			select {
			case <-r.Context().Done():
			case <-time.After(stall * 30):
			}
			return
		}
		end := int64(len(payload)) - 1
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(payload)))
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(payload))-start, 10))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload[start:])
	}))
	t.Cleanup(srv.Close)

	svc := NewMarketService(MarketConfig{
		IndexURL:     srv.URL,
		PluginDir:    t.TempDir(),
		StallTimeout: stall,
	})
	dest := filepath.Join(svc.pluginDir, "test.zip")
	if err := svc.fetchZip(context.Background(), srv.URL, dest, nil); err != nil {
		t.Fatalf("fetchZip failed: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("resumed download produced wrong content: got %d bytes, want %d", len(got), len(payload))
	}
	if attempts.Load() < 2 {
		t.Fatalf("expected at least 2 attempts (one stalled, one resumed), got %d", attempts.Load())
	}
}

func TestIsRetryableDownloadError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"generic", errors.New("boom"), true},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"wrapped deadline", fmt.Errorf("fetch: %w", context.DeadlineExceeded), true},
		{"stalled", errDownloadStalled, true},
		{"wrapped stalled", fmt.Errorf("market: %w", errDownloadStalled), true},
		{"canceled", context.Canceled, false},
		{"wrapped canceled", fmt.Errorf("ctx: %w", context.Canceled), false},
		{"non-retryable", errDownloadNotRetryable, false},
		{"wrapped non-retryable", fmt.Errorf("size: %w", errDownloadNotRetryable), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isRetryableDownloadError(c.err); got != c.want {
				t.Fatalf("isRetryableDownloadError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestJob_SetProgressMonotonicTotal(t *testing.T) {
	job := newInstallJob("alpha")
	job.setProgress(10, 100)
	job.setProgress(50, 80)
	st := job.Status()
	if st.Downloaded != 50 {
		t.Fatalf("downloaded: want 50, got %d", st.Downloaded)
	}
	if st.Total != 100 {
		t.Fatalf("total should stay at the larger value 100, got %d", st.Total)
	}
	job.setProgress(60, -1)
	if job.Status().Total != 100 {
		t.Fatalf("total should remain 100 after a -1 sample, got %d", job.Status().Total)
	}
}

func TestListMarket_StatusRecomputedFromCache(t *testing.T) {
	var hits atomic.Int32
	body := indexJSON(t, entry("alpha", "1.0.0", releaseWithCurrentAsset()))
	svc := newTestService(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		io.WriteString(w, body)
	})

	first, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first[0].Status != "not_installed" {
		t.Fatalf("first listing: want not_installed, got %s", first[0].Status)
	}

	installed := filepath.Join(svc.pluginDir, "alpha")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	metaBytes, _ := json.Marshal(plugin.NewMarketMeta("official", "1.0.0", "1.0.0", "alpha/v1.0.0"))
	if err := os.WriteFile(filepath.Join(installed, ".market.json"), metaBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	second, err := svc.ListMarket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if second[0].Status != "installed" {
		t.Fatalf("status must be recomputed from the warm cache after install: want installed, got %s", second[0].Status)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("index server should be hit only once while cache is warm, got %d", got)
	}
}
