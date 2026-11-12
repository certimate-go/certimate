package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var validAccessUsages = map[string]bool{
	"dns":          true,
	"hosting":      true,
	"notification": true,
	"ca":           true,
}

const (
	CapabilityDeploy = "deploy"
	CapabilityDns    = "dns"
	CapabilityNotify = "notify"
)

var validCapabilities = map[string]bool{
	CapabilityDeploy: true,
	CapabilityDns:    true,
	CapabilityNotify: true,
}

func IsValidCapability(capability string) bool {
	return validCapabilities[capability]
}

type Manifest struct {
	Version            string   `json:"version"`
	ProviderType       string   `json:"provider_type"`
	AccessProviderType string   `json:"access_provider_type"`
	Capability         string   `json:"capability"`
	DisplayNameKey     string   `json:"display_name_key"`
	DeployCategory     string   `json:"deploy_category"`
	ProtocolVersion    uint32   `json:"protocol_version"`
	MinCoreVersion     string   `json:"min_core_version"`
	MaxCoreVersion     string   `json:"max_core_version"`
	OS                 string   `json:"os"`
	Arch               string   `json:"arch"`
	Binary             string   `json:"binary"`
	Icon               string   `json:"icon"`
	SHA256             string   `json:"sha256"`
	Usages             []string `json:"usages"`
	Priority           int      `json:"priority"`
	Description        string   `json:"description"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest json: %w", err)
	}
	capability, err := parseCapabilityField(fields["capability"])
	if err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest json: %w", err)
	}
	delete(fields, "capability")
	rest, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest json: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(rest, &m); err != nil {
		return nil, fmt.Errorf("plugin: invalid manifest json: %w", err)
	}
	m.Capability = capability
	return &m, nil
}

func parseCapabilityField(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single, nil
	}
	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err == nil {
		if len(multiple) == 1 {
			return multiple[0], nil
		}
		return "", fmt.Errorf("capability must declare exactly one of deploy, dns, notify (got %d)", len(multiple))
	}
	return "", fmt.Errorf("capability must be a string or a single-element array")
}

func (m *Manifest) Validate() error {
	var problems []string
	if m.ProviderType == "" {
		problems = append(problems, "provider_type must not be empty")
	}
	if m.AccessProviderType == "" {
		problems = append(problems, "access_provider_type must not be empty")
	}
	if !validCapabilities[m.Capability] {
		problems = append(problems, fmt.Sprintf("capability must be one of deploy, dns, notify (got %q)", m.Capability))
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
	for _, u := range m.Usages {
		if !validAccessUsages[u] {
			problems = append(problems, fmt.Sprintf("usages contains invalid value %q", u))
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
