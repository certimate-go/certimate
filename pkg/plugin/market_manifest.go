package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	providerTypePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)
	repoPattern         = regexp.MustCompile(`^certimate-go/[a-z0-9_.-]+$`)
)

type Release struct {
	Repo      string            `json:"repo"`
	Tag       string            `json:"tag"`
	Assets    map[string]string `json:"assets"`
	Checksums map[string]string `json:"checksums,omitempty"`
}

type MarketManifest struct {
	Manifest
	Release *Release `json:"release,omitempty"`
}

func ParseMarketManifest(data []byte) (*MarketManifest, error) {
	var m MarketManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("plugin: invalid market manifest json: %w", err)
	}
	if m.Release == nil {
		return nil, fmt.Errorf("plugin: market manifest for %q has no release block", m.ProviderType)
	}
	return &m, nil
}

func ValidateProviderType(pt string) error {
	if pt == "" {
		return fmt.Errorf("plugin: provider_type must not be empty")
	}
	if !providerTypePattern.MatchString(pt) {
		return fmt.Errorf("plugin: provider_type %q contains invalid characters (allowed: a-z, 0-9, underscore, dot, hyphen)", pt)
	}
	return nil
}

func ValidateBinaryName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin: binary must not be empty")
	}
	if !providerTypePattern.MatchString(name) {
		return fmt.Errorf("plugin: binary %q contains invalid characters (allowed: a-z, 0-9, underscore, dot, hyphen)", name)
	}
	return nil
}

func ValidateReleaseRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("plugin: release repo must not be empty")
	}
	if !repoPattern.MatchString(repo) {
		return fmt.Errorf("plugin: release repo %q is not in the trusted organization (certimate-go/)", repo)
	}
	return nil
}

func AssetKey(goos, goarch string) string {
	return goos + "/" + goarch
}
