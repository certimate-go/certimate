package pluginhost

import (
	"testing"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestCatalog_AddAndEntries(t *testing.T) {
	c := NewCatalog()
	c.Add(&CatalogEntry{Source: SourcePlugin, ProviderType: "b", Category: "other"})
	c.Add(&CatalogEntry{Source: SourcePlugin, ProviderType: "a", Category: "other"})

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

func TestCatalog_PurgeFactoryRegistersAndUnregisters(t *testing.T) {
	const name = "catalog-purge-test"
	catalog := NewCatalog()
	t.Cleanup(catalog.Close)
	factory := &pluginPurgerFactory{}
	p := &registeredPlugin{
		source:      &plugin.DiscoveredPlugin{Manifest: &plugin.Manifest{ProviderType: name, Capability: plugin.CapabilityPurge}},
		capability:  plugin.CapabilityPurge,
		factory:     factory,
		newPurger:   factory.new,
		entry:       &CatalogEntry{ProviderType: name},
	}
	if err := catalog.publish(p); err != nil {
		t.Fatal(err)
	}
	if _, err := purgers.Registries.Get(domain.PurgeProviderType(name)); err != nil {
		t.Fatalf("purge factory not registered: %v", err)
	}
	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(name)); err == nil {
		t.Fatal("purge plugin must not register as deployer")
	}
	catalog.unregister(name)
	if _, err := purgers.Registries.Get(domain.PurgeProviderType(name)); err == nil {
		t.Fatal("purge factory survived unregister")
	}
}
