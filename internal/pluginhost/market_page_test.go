package pluginhost

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestMarketIconURL_UsesJsDelivrForGitHubRaw(t *testing.T) {
	svc := NewMarketService(MarketConfig{
		IndexURL:  "https://raw.githubusercontent.com/certimate-go/plugins/refs/heads/main/index.json",
		PluginDir: t.TempDir(),
	})
	got := svc.marketIconURL("panel1/1panel.svg")
	want := "https://cdn.jsdelivr.net/gh/certimate-go/plugins@main/panel1/1panel.svg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMarketIconURL_UsesJsDelivrForShortGitHubRaw(t *testing.T) {
	svc := NewMarketService(MarketConfig{
		IndexURL:  "https://raw.githubusercontent.com/certimate-go/plugins/main/index.json",
		PluginDir: t.TempDir(),
	})
	got := svc.marketIconURL("panel1/1panel.svg")
	want := "https://cdn.jsdelivr.net/gh/certimate-go/plugins@main/panel1/1panel.svg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMarketIconURL_ResolvesRelativeForCustomIndex(t *testing.T) {
	svc := NewMarketService(MarketConfig{
		IndexURL:  "https://example.test/market/index.json",
		PluginDir: t.TempDir(),
	})
	got := svc.marketIconURL("panel1/1panel.svg")
	want := "https://example.test/market/panel1/1panel.svg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMarketIconURL_RejectsInvalidPath(t *testing.T) {
	svc := NewMarketService(MarketConfig{
		IndexURL:  "https://raw.githubusercontent.com/certimate-go/plugins/main/index.json",
		PluginDir: t.TempDir(),
	})
	if got := svc.marketIconURL("../escape.svg"); got != "" {
		t.Fatalf("expected empty for path escape, got %q", got)
	}
	if got := svc.marketIconURL(""); got != "" {
		t.Fatalf("expected empty for empty path, got %q", got)
	}
}

func TestListMarketPage_CapabilityFilter(t *testing.T) {
	deployEntry := entry("deploy-one", "1.0.0", nil)
	deployEntry["capability"] = "deploy"
	dnsEntry := entry("dns-one", "1.0.0", nil)
	dnsEntry["capability"] = "dns"
	notifyEntry := entry("notify-one", "1.0.0", nil)
	notifyEntry["capability"] = "notify"
	body := indexJSON(t, deployEntry, dnsEntry, notifyEntry, entry("plain-one", "1.0.0", nil))
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	})

	list := func(t *testing.T, query MarketQuery) *MarketPage {
		t.Helper()
		page, err := svc.ListMarketPage(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	capabilityOf := func(t *testing.T, page *MarketPage, providerType string) string {
		t.Helper()
		for _, item := range page.Items {
			if item.ProviderType == providerType {
				return item.Capability
			}
		}
		t.Fatalf("provider %q missing from listing", providerType)
		return ""
	}

	unfiltered := list(t, MarketQuery{Page: 1, PerPage: 30})
	if len(unfiltered.Items) != 4 {
		t.Fatalf("want 4 items without capability filter, got %d", len(unfiltered.Items))
	}
	if got := capabilityOf(t, unfiltered, "plain-one"); got != plugin.CapabilityDeploy {
		t.Fatalf("want missing capability defaulted to deploy, got %q", got)
	}
	if got := list(t, MarketQuery{Page: 1, PerPage: 30, Capability: "all"}); len(got.Items) != 4 {
		t.Fatalf("want 4 items for capability all, got %d", len(got.Items))
	}
	got := list(t, MarketQuery{Page: 1, PerPage: 30, Capability: "dns"})
	if len(got.Items) != 1 || got.Items[0].ProviderType != "dns-one" || got.Items[0].Capability != "dns" {
		t.Fatalf("want exactly the dns provider, got %+v", got.Items)
	}
	if got := list(t, MarketQuery{Page: 1, PerPage: 30, Capability: plugin.CapabilityDeploy}); len(got.Items) != 2 {
		t.Fatalf("want 2 items for capability deploy (explicit and defaulted), got %d", len(got.Items))
	}
	combined := list(t, MarketQuery{Page: 1, PerPage: 30, Capability: plugin.CapabilityDeploy, Keyword: "dns-one"})
	if len(combined.Items) != 0 {
		t.Fatalf("want 0 items when capability and keyword filters compose, got %d", len(combined.Items))
	}
}

func TestMarketTranslationInstalledFallback(t *testing.T) {
	manifest := &plugin.MarketManifest{Manifest: plugin.Manifest{
		ProviderType: "legacy-provider", DisplayNameKey: "plugin.legacy-provider.name", Description: "Legacy description",
	}}
	installed := &registeredPlugin{entry: &CatalogEntry{
		DisplayNameKey: "plugin.legacy-provider.name",
		I18n: map[string]map[string]string{
			"zh": {"plugin.legacy-provider.name": "已安装的插件", "plugin.legacy-provider.description": "已安装的描述"},
			"en": {"plugin.legacy-provider.name": "Installed plugin"},
		},
	}}
	for _, check := range []struct {
		locale    string
		installed *registeredPlugin
		want      plugin.MarketTranslation
	}{
		{"zh-CN", installed, plugin.MarketTranslation{Name: "已安装的插件", Description: "已安装的描述"}},
		{"fr", installed, plugin.MarketTranslation{Name: "Installed plugin", Description: "Legacy description"}},
		{"zh", nil, plugin.MarketTranslation{Name: "legacy-provider", Description: "Legacy description"}},
	} {
		if got := marketTranslation(manifest, check.installed, check.locale); got != check.want {
			t.Errorf("locale %q: got %+v, want %+v", check.locale, got, check.want)
		}
	}
	manifest.I18n = map[string]plugin.MarketTranslation{"zh": {Name: "市场名称", Description: "市场描述"}}
	if got := marketTranslation(manifest, installed, "zh-CN"); got != manifest.I18n["zh"] {
		t.Errorf("catalog translations must take precedence: %+v", got)
	}
}
