package plugin

import "testing"

func TestMarketSourceIdentity(t *testing.T) {
	a, err := ParseMarketSource("https://github.com/Example/Plugins.git/")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseMarketSource("https://github.com/example/plugins")
	if err != nil || a != b || a.ID == "official" {
		t.Fatalf("canonical sources: %+v %+v %v", a, b, err)
	}
	for _, raw := range []string{"http://github.com/a/b", "https://github.com/a/b/tree/main", "https://github.com/a/b?q=x", "https://user@github.com/a/b", "https://github.com:443/a/b", "https://github.com/a/%62", "https://github.com/a/..", "https://github.com/a/b.git.git", "https://example.com/a/b"} {
		if _, err := ParseMarketSource(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	official, err := ParseMarketSource("https://github.com/Certimate-Go/Plugins.git")
	if err != nil || official != OfficialMarketSource() {
		t.Fatalf("official alias: %+v %v", official, err)
	}
}

func TestMarketReceiptOrigin(t *testing.T) {
	custom, _ := ParseMarketSource("https://github.com/example/plugins")
	for _, tc := range []struct {
		meta *MarketMeta
		want string
	}{
		{nil, ""},
		{&MarketMeta{Source: "official"}, "official"},
		{&MarketMeta{Source: "unknown"}, ""},
		{&MarketMeta{Source: custom.ID, Market: &custom}, custom.ID},
		{&MarketMeta{Source: "official", Market: &custom}, ""},
		{&MarketMeta{Source: custom.ID}, ""},
	} {
		if got := tc.meta.Origin().ID; got != tc.want {
			t.Errorf("origin %v = %q, want %q", tc.meta, got, tc.want)
		}
	}
}
