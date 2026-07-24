package pluginhost

import (
	"sort"
	"sync"
)

const SourcePlugin = "plugin"

type CatalogEntry struct {
	Source               string                       `json:"source"`
	ProviderType         string                       `json:"providerType"`
	AccessProviderType   string                       `json:"accessProviderType"`
	DeployCategory       string                       `json:"deployCategory"`
	DisplayNameKey       string                       `json:"displayNameKey"`
	AccessDisplayNameKey string                       `json:"accessDisplayNameKey"`
	Icon                 string                       `json:"icon,omitempty"`
	I18n                 map[string]map[string]string `json:"i18n,omitempty"`
}

type Catalog struct {
	mu      sync.RWMutex
	entries []*CatalogEntry
}

func NewCatalog() *Catalog {
	return &Catalog{}
}

func (c *Catalog) Add(entry *CatalogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entry)
	c.sortLocked()
}

func (c *Catalog) sortLocked() {
	sort.SliceStable(c.entries, func(i, j int) bool {
		return c.entries[i].ProviderType < c.entries[j].ProviderType
	})
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
