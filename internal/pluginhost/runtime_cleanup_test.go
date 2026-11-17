package pluginhost

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/plugin"
)

func TestRuntimeDirectoryFollowsFactoryLifetime(t *testing.T) {
	const name = "runtime-cleanup-test"
	catalog := NewCatalog()
	t.Cleanup(catalog.Close)
	dir := t.TempDir()
	publish := func() deployers.ProviderFactoryFunc {
		t.Helper()
		factory := &pluginDeployerFactory{}
		p := &registeredPlugin{
			source:      &plugin.DiscoveredPlugin{Manifest: &plugin.Manifest{ProviderType: name, Capability: plugin.CapabilityDeploy}},
			runtimeDir:  dir,
			capability:  plugin.CapabilityDeploy,
			factory:     factory,
			newDeployer: factory.new,
			entry:       &CatalogEntry{ProviderType: name},
		}
		if err := catalog.publish(p); err != nil {
			t.Fatal(err)
		}
		factoryFn, err := deployers.Registries.Get(domain.DeploymentProviderType(name))
		if err != nil {
			t.Fatal(err)
		}
		return factoryFn
	}
	factory := publish()
	catalog.unregister(name)
	runtime.GC()
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("held factory lost runtime directory: %v", err)
	}
	runtime.KeepAlive(factory)
	factory = nil
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("unreferenced factory retained runtime directory")
}

func TestRuntimeDirectoryFollowsPurgerFactoryLifetime(t *testing.T) {
	const name = "runtime-cleanup-purge-test"
	catalog := NewCatalog()
	t.Cleanup(catalog.Close)
	dir := t.TempDir()
	factory := &pluginPurgerFactory{}
	p := &registeredPlugin{
		source:      &plugin.DiscoveredPlugin{Manifest: &plugin.Manifest{ProviderType: name, Capability: plugin.CapabilityPurge}},
		runtimeDir:  dir,
		capability:  plugin.CapabilityPurge,
		factory:     factory,
		newPurger:   factory.new,
		entry:       &CatalogEntry{ProviderType: name},
	}
	if err := catalog.publish(p); err != nil {
		t.Fatal(err)
	}
	factoryFn, err := purgers.Registries.Get(domain.PurgeProviderType(name))
	if err != nil {
		t.Fatal(err)
	}
	catalog.unregister(name)
	runtime.GC()
	runtime.KeepAlive(factoryFn)
	factoryFn = nil
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("unreferenced purger factory retained runtime directory")
}
