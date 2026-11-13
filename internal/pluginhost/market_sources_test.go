package pluginhost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/repository"
	_ "github.com/certimate-go/certimate/migrations"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type sourceRoundTripper func(*http.Request) (*http.Response, error)

func (f sourceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestMarketSourceRefreshCannotOverwriteReaddedSource(t *testing.T) {
	ctx := context.Background()
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: t.TempDir()})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	defer app.ResetBootstrapState()
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	cache := repository.NewPluginMarketCache(app)
	started, resume := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	svc := NewMarketService(MarketConfig{
		Sources: repository.NewPluginMarketSources(app), Cache: cache, PluginDir: t.TempDir(), CacheTTL: time.Hour,
		CatalogTransport: sourceRoundTripper(func(r *http.Request) (*http.Response, error) {
			body := `{"plugins":[{"provider_type":"new","version":"2.0.0"}]}`
			if calls.Add(1) == 2 {
				close(started)
				<-resume
				body = `{"plugins":[{"provider_type":"old","version":"1.0.0"}]}`
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		}),
	})
	source, _ := plugin.ParseMarketSource("https://github.com/example/market")
	if _, err := svc.AddSource(ctx, "https://github.com/example/market"); err != nil {
		t.Fatal(err)
	}
	pending := make(chan error, 1)
	go func() { _, err := svc.RefreshSource(ctx, source.ID); pending <- err }()
	defer func() {
		select {
		case <-resume:
		default:
			close(resume)
		}
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("refresh did not start")
	}
	if err := svc.RemoveSource(ctx, source.ID); err != nil {
		t.Fatal(err)
	}
	if status, err := svc.AddSource(ctx, "https://github.com/example/market"); err != nil || status.ID != source.ID || status.Error != "" {
		t.Fatalf("re-add: %+v %v", status, err)
	}
	close(resume)
	if err := <-pending; err != ErrMarketUnavailable {
		t.Fatalf("old refresh result: %v", err)
	}
	cached, _, err := cache.Load(ctx, source.IndexURL())
	if err != nil || len(cached) != 1 || cached[0].ProviderType != "new" {
		t.Fatalf("old generation changed the persistent cache: %v %v", cached, err)
	}
	page, err := svc.ListMarketPage(ctx, MarketQuery{Source: source.ID, PerPage: 10})
	if err != nil || len(page.Items) != 1 || page.Items[0].ProviderType != "new" {
		t.Fatalf("wrong live catalog: %+v %v", page, err)
	}
}

func TestMarketRejectsCatalogRedirectOutsideSourceOrigin(t *testing.T) {
	var reachedOtherHost bool
	svc := NewMarketService(MarketConfig{
		CatalogTransport: sourceRoundTripper(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host != "raw.githubusercontent.com" {
				reachedOtherHost = true
			}
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://example.test/elsewhere"}}, Body: http.NoBody}, nil
		}),
	})
	status, err := svc.RefreshSource(context.Background(), "official")
	if err != nil || reachedOtherHost || !strings.Contains(status.Error, "unexpected catalog redirect") {
		t.Fatalf("external redirect: %+v %v, reached=%t", status, err, reachedOtherHost)
	}
}

func TestMarketReleaseRedirectUsesGitHubAssetOrigin(t *testing.T) {
	for _, target := range []struct {
		host string
		want bool
	}{
		{"release-assets.githubusercontent.com", true},
		{"example.test", false},
	} {
		t.Run(target.host, func(t *testing.T) {
			var followed bool
			svc := NewMarketService(MarketConfig{DownloadTransport: sourceRoundTripper(func(req *http.Request) (*http.Response, error) {
				if req.URL.Host == "github.com" {
					return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://" + target.host + "/asset.zip"}}, Body: http.NoBody}, nil
				}
				followed = true
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("release bytes")), ContentLength: 13}, nil
			})})
			err := svc.downloadAttempt(context.Background(), "https://github.com/fixture/releases/releases/download/v1/plugin.zip", t.TempDir()+"/asset.zip", nil)
			if followed != target.want || (err == nil) != target.want {
				t.Fatalf("redirect followed=%t, error=%v", followed, err)
			}
		})
	}
}
