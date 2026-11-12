package pluginhost

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/certimate-go/certimate/internal/certacme/certifiers"
	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/notify/notifiers"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/pkg/plugin"
)

const (
	SourcePlugin      = "plugin"
	SourceMarketplace = "marketplace"
)

var ErrAccessOwnerInUse = errors.New("pluginhost: access owner still in use")

type CatalogEntry struct {
	Source               string                       `json:"source"`
	Capability           string                       `json:"capability"`
	ProviderType         string                       `json:"providerType"`
	AccessProviderType   string                       `json:"accessProviderType"`
	OwnsAccess           bool                         `json:"ownsAccess"`
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
		if e.Capability == plugin.CapabilityDeploy {
			byAccess[e.AccessProviderType] = append(byAccess[e.AccessProviderType], e.ProviderType)
		}
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

func (c *Catalog) resolveAccessOwnership(prepared []*registeredPlugin) ([]*registeredPlugin, []error) {
	registeredOwners := c.accessOwners()

	sorted := make([]*registeredPlugin, len(prepared))
	copy(sorted, prepared)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].source.Manifest.ProviderType < sorted[j].source.Manifest.ProviderType
	})

	owners := make(map[string]string)
	var errs []error
	failed := make(map[*registeredPlugin]bool)
	for _, p := range sorted {
		if !p.ownsAccess {
			continue
		}
		name := p.source.Manifest.ProviderType
		if existing, ok := owners[p.accessType]; ok && existing != name {
			errs = append(errs, fmt.Errorf("plugin %q and plugin %q both claim ownership of access type %q; keep exactly one access schema", existing, name, p.accessType))
			failed[p] = true
			continue
		}
		if existing, ok := registeredOwners[p.accessType]; ok && existing != name {
			errs = append(errs, fmt.Errorf("plugin %q and plugin %q both claim ownership of access type %q; keep exactly one access schema", existing, name, p.accessType))
			failed[p] = true
			continue
		}
		owners[p.accessType] = name
	}
	for _, p := range sorted {
		if p.ownsAccess || failed[p] {
			continue
		}
		accessType := p.accessType
		name := p.source.Manifest.ProviderType
		_, preparedOwned := owners[accessType]
		registeredOwner, registeredOwned := registeredOwners[accessType]
		if preparedOwned || (registeredOwned && registeredOwner != name) || isCoreOwnedAccessType(accessType) {
			continue
		}
		errs = append(errs, fmt.Errorf("plugin %q reuses access type %q but no owner is installed and it is not a built-in; install the access owner plugin first", name, accessType))
		failed[p] = true
	}

	ordered := make([]*registeredPlugin, 0, len(sorted))
	for _, p := range sorted {
		if !failed[p] && p.ownsAccess {
			ordered = append(ordered, p)
		}
	}
	for _, p := range sorted {
		if !failed[p] && !p.ownsAccess {
			ordered = append(ordered, p)
		}
	}
	return ordered, errs
}

func (c *Catalog) accessOwners() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	owners := make(map[string]string, len(c.plugins))
	for name, p := range c.plugins {
		if p.ownsAccess {
			owners[p.accessType] = name
		}
	}
	return owners
}

func (c *Catalog) accessReusers(accessType string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var reusers []string
	for name, p := range c.plugins {
		if !p.ownsAccess && p.accessType == accessType {
			reusers = append(reusers, name)
		}
	}
	sort.Strings(reusers)
	return reusers
}

func (c *Catalog) refuseRemoval(name string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.plugins[name]
	if !ok || !p.ownsAccess {
		return nil
	}
	var reusers []string
	for other, q := range c.plugins {
		if other != name && !q.ownsAccess && q.accessType == p.accessType {
			reusers = append(reusers, other)
		}
	}
	if len(reusers) == 0 {
		return nil
	}
	sort.Strings(reusers)
	return fmt.Errorf("%w: plugin %q owns access type %q still used by %s; remove them first", ErrAccessOwnerInUse, name, p.accessType, strings.Join(reusers, ", "))
}

func (c *Catalog) publish(p *registeredPlugin) error {
	name := p.source.Manifest.ProviderType
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.registerFactoryLocked(name, p); err != nil {
		return err
	}
	providerschema.Envelopes.ReplaceOwner(c.owner+":"+name, p.envelopes)
	c.plugins[name] = p
	c.runtimeDirs.add(p.runtimeDir)
	c.addRuntimeCleanup(p)
	c.removeEntryLocked(name)
	c.entries = append(c.entries, p.entry)
	c.sortLocked()
	c.buildDeployersLocked()
	return nil
}

func (c *Catalog) addRuntimeCleanup(p *registeredPlugin) {
	dir := runtimeDirectory{set: c.runtimeDirs, path: p.runtimeDir}
	switch f := p.factory.(type) {
	case *pluginDeployerFactory:
		runtime.AddCleanup(f, removeRuntimeDirectory, dir)
	case *pluginDnsChallengerFactory:
		runtime.AddCleanup(f, removeRuntimeDirectory, dir)
	case *pluginNotifierFactory:
		runtime.AddCleanup(f, removeRuntimeDirectory, dir)
	}
}

func (c *Catalog) registerFactoryLocked(name string, p *registeredPlugin) error {
	switch p.capability {
	case plugin.CapabilityDns:
		return certifiers.ACMEDns01Registries.SetOwned(domain.ACMEDns01ProviderType(name), c.owner, p.newChallenger)
	case plugin.CapabilityNotify:
		return notifiers.Registries.SetOwned(domain.NotificationProviderType(name), c.owner, p.newNotifier)
	default:
		return deployers.Registries.SetOwned(domain.DeploymentProviderType(name), c.owner, p.newDeployer)
	}
}

func (c *Catalog) unregisterFactoryLocked(name string, p *registeredPlugin) {
	switch p.capability {
	case plugin.CapabilityDns:
		certifiers.ACMEDns01Registries.RemoveOwned(domain.ACMEDns01ProviderType(name), c.owner)
	case plugin.CapabilityNotify:
		notifiers.Registries.RemoveOwned(domain.NotificationProviderType(name), c.owner)
	default:
		deployers.Registries.RemoveOwned(domain.DeploymentProviderType(name), c.owner)
	}
}

func (c *Catalog) unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p, ok := c.plugins[name]
	if !ok {
		return
	}
	c.unregisterFactoryLocked(name, p)
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
