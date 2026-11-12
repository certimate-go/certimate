package pluginhost

import (
	"fmt"
	"maps"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

const (
	SourcePlugin      = "plugin"
	SourceMarketplace = "marketplace"
)

type CatalogEntry struct {
	Source               string                       `json:"source"`
	ProviderType         string                       `json:"providerType"`
	AccessProviderType   string                       `json:"accessProviderType"`
	DeployCategory       string                       `json:"deployCategory"`
	DisplayNameKey       string                       `json:"displayNameKey"`
	AccessDisplayNameKey string                       `json:"accessDisplayNameKey"`
	Icon                 string                       `json:"icon,omitempty"`
	I18n                 map[string]map[string]string `json:"i18n,omitempty"`
	AccessUsages         []string                     `json:"accessUsages,omitempty"`
	Priority             int                          `json:"priority"`
	Description          string                       `json:"description,omitempty"`
	Deployers            []string                     `json:"deployers,omitempty"`
}

type Catalog struct {
	mu          sync.RWMutex
	entries     []*CatalogEntry
	plugins     map[string]*registeredPlugin
	owner       string
	runtimeDirs *runtimeDirectorySet
}

var catalogSequence atomic.Uint64

func NewCatalog() *Catalog {
	return &Catalog{plugins: make(map[string]*registeredPlugin), owner: fmt.Sprintf("plugin-catalog-%d", catalogSequence.Add(1)), runtimeDirs: &runtimeDirectorySet{dirs: make(map[string]struct{})}}
}

func (c *Catalog) Add(entry *CatalogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, existing := range c.entries {
		if existing.ProviderType == entry.ProviderType {
			c.entries[i] = entry
			c.sortLocked()
			c.buildDeployersLocked()
			return
		}
	}
	c.entries = append(c.entries, entry)
	c.sortLocked()
	c.buildDeployersLocked()
}

func (c *Catalog) sortLocked() {
	sort.SliceStable(c.entries, func(i, j int) bool {
		if c.entries[i].Priority != c.entries[j].Priority {
			return c.entries[i].Priority > c.entries[j].Priority
		}
		return c.entries[i].ProviderType < c.entries[j].ProviderType
	})
}

func (c *Catalog) buildDeployersLocked() {
	byAccess := make(map[string][]string)
	for _, e := range c.entries {
		byAccess[e.AccessProviderType] = append(byAccess[e.AccessProviderType], e.ProviderType)
	}
	for _, e := range c.entries {
		e.Deployers = byAccess[e.ProviderType]
	}
}

func (c *Catalog) Entries() []CatalogEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]CatalogEntry, len(c.entries))
	for i, e := range c.entries {
		out[i] = *e
	}
	return out
}

func (c *Catalog) Remove(providerType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.removeEntryLocked(providerType) {
		c.buildDeployersLocked()
	}
}

func (c *Catalog) removeEntryLocked(providerType string) bool {
	for i, entry := range c.entries {
		if entry.ProviderType == providerType {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Catalog) I18nBundles() map[string]map[string]map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]map[string]map[string]string, len(c.entries))
	for _, e := range c.entries {
		if len(e.I18n) == 0 {
			continue
		}
		out[e.ProviderType] = e.I18n
	}
	return out
}

func (c *Catalog) registered() map[string]*registeredPlugin {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]*registeredPlugin, len(c.plugins))
	maps.Copy(out, c.plugins)
	return out
}

func (c *Catalog) publish(p *registeredPlugin) error {
	name := p.source.Manifest.ProviderType
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := deployers.Registries.SetOwned(domain.DeploymentProviderType(name), c.owner, p.factory.new); err != nil {
		return err
	}
	providerschema.Envelopes.ReplaceOwner(c.owner+":"+name, p.envelopes)
	c.plugins[name] = p
	c.runtimeDirs.add(p.runtimeDir)
	runtime.AddCleanup(p.factory, removeRuntimeDirectory, runtimeDirectory{set: c.runtimeDirs, path: p.runtimeDir})
	c.removeEntryLocked(name)
	c.entries = append(c.entries, p.entry)
	c.sortLocked()
	c.buildDeployersLocked()
	return nil
}

func (c *Catalog) unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.plugins[name]; !ok {
		return
	}
	deployers.Registries.RemoveOwned(domain.DeploymentProviderType(name), c.owner)
	providerschema.Envelopes.ReplaceOwner(c.owner+":"+name, nil)
	delete(c.plugins, name)
	c.removeEntryLocked(name)
	c.buildDeployersLocked()
}

func (c *Catalog) Close() {
	for name := range c.registered() {
		c.unregister(name)
	}
	c.runtimeDirs.close()
}

type runtimeDirectorySet struct {
	mu   sync.Mutex
	dirs map[string]struct{}
}

type runtimeDirectory struct {
	set  *runtimeDirectorySet
	path string
}

func (s *runtimeDirectorySet) add(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirs[dir] = struct{}{}
}

func (s *runtimeDirectorySet) remove(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.RemoveAll(dir); err == nil {
		delete(s.dirs, dir)
	}
}

func (s *runtimeDirectorySet) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for dir := range s.dirs {
		if err := os.RemoveAll(dir); err == nil {
			delete(s.dirs, dir)
		}
	}
}

func removeRuntimeDirectory(dir runtimeDirectory) {
	dir.set.remove(dir.path)
}
