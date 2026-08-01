package pluginhost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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
