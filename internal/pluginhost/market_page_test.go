package pluginhost

import "testing"

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
