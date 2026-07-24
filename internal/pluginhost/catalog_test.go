package pluginhost

import (
	"testing"
)

func TestCatalog_AddAndEntries(t *testing.T) {
	c := NewCatalog()
	c.Add(&CatalogEntry{Source: SourcePlugin, ProviderType: "b", DeployCategory: "other"})
	c.Add(&CatalogEntry{Source: SourcePlugin, ProviderType: "a", DeployCategory: "other"})

	got := c.Entries()
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].ProviderType != "a" || got[1].ProviderType != "b" {
		t.Fatalf("entries not sorted by type: %+v", got)
	}
}

func TestCatalog_I18nBundles(t *testing.T) {
	c := NewCatalog()
	c.Add(&CatalogEntry{
		ProviderType: "demo",
		I18n:         map[string]map[string]string{"zh": {"plugin.demo.name": "示例"}},
	})
	c.Add(&CatalogEntry{ProviderType: "noi18n"})

	bundles := c.I18nBundles()
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if bundles["demo"]["zh"]["plugin.demo.name"] != "示例" {
		t.Fatalf("bundle content mismatch: %+v", bundles)
	}
}
