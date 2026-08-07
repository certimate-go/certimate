package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MarketMeta records provenance and version information for a marketplace-installed
// plugin. Stored as .market.json alongside the plugin's manifest.json.
type MarketMeta struct {
	Source                 string `json:"source"`
	InstalledVersion       string `json:"installed_version"`
	MarketVersionAtInstall string `json:"market_version_at_install"`
	InstalledAt            string `json:"installed_at"`
	ReleaseTag             string `json:"release_tag"`
}

// ReadMarketMeta reads the .market.json file for a plugin in the given directory.
// Returns (nil, nil) if the file does not exist (plugin is not marketplace-managed).
func ReadMarketMeta(pluginDir, providerType string) (*MarketMeta, error) {
	path := filepath.Join(pluginDir, providerType, ".market.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("plugin: read market meta for %q: %w", providerType, err)
	}
	var meta MarketMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("plugin: parse market meta for %q: %w", providerType, err)
	}
	return &meta, nil
}

// WriteMarketMeta writes the .market.json file for a plugin.
func WriteMarketMeta(pluginDir, providerType string, meta *MarketMeta) error {
	path := filepath.Join(pluginDir, providerType, ".market.json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("plugin: marshal market meta for %q: %w", providerType, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("plugin: write market meta for %q: %w", providerType, err)
	}
	return nil
}

// NewMarketMeta creates a MarketMeta with the current timestamp.
func NewMarketMeta(source, installedVersion, marketVersion, releaseTag string) *MarketMeta {
	return &MarketMeta{
		Source:                 source,
		InstalledVersion:       installedVersion,
		MarketVersionAtInstall: marketVersion,
		InstalledAt:            time.Now().UTC().Format(time.RFC3339),
		ReleaseTag:             releaseTag,
	}
}

// CompareVersions compares two version strings using semver comparison.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Falls back to string comparison for non-semver versions.
func CompareVersions(a, b string) int {
	return semverCompare(a, b)
}
