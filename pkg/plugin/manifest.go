package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Manifest struct {
	Version            string `json:"version"`
	ProviderType       string `json:"provider_type"`
	AccessProviderType string `json:"access_provider_type"`
	DisplayNameKey     string `json:"display_name_key"`
	DeployCategory     string `json:"deploy_category"`
	ProtocolVersion    uint32 `json:"protocol_version"`
	MinCoreVersion     string `json:"min_core_version"`
	MaxCoreVersion     string `json:"max_core_version"`
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	Binary             string `json:"binary"`
	Icon               string `json:"icon"`
	SHA256             string `json:"sha256"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest json: %w", err)
	}
	return &m, nil
}

func (m *Manifest) Validate() error {
	var problems []string
	if m.ProviderType == "" {
		problems = append(problems, "provider_type must not be empty")
	}
	if m.AccessProviderType == "" {
		problems = append(problems, "access_provider_type must not be empty")
	}
	if m.Binary == "" {
		problems = append(problems, "binary must not be empty")
	}
	if m.ProtocolVersion == 0 {
		problems = append(problems, "protocol_version must not be zero")
	}
	if m.OS != "" || m.Arch != "" {
		if m.OS == "" || m.Arch == "" {
			problems = append(problems, "os and arch must be specified together")
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("plugin: manifest for %q invalid: %s", m.ProviderType, strings.Join(problems, "; "))
	}
	return nil
}

func loadManifest(dir string) (*Manifest, []byte, error) {
	path := dir + "/manifest.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	m, err := ParseManifest(data)
	if err != nil {
		return nil, nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, nil, err
	}
	return m, data, nil
}
