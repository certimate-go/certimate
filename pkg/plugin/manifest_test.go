package plugin

import (
	"testing"
)

func TestParseManifest_Valid(t *testing.T) {
	raw := []byte(`{
		"version": "0.1.0",
		"provider_type": "webhook-deployer",
		"access_provider_type": "webhook-deployer",
		"capability": "deploy",
		"display_name_key": "plugin.webhook-deployer.name",
		"deploy_category": "other",
		"protocol_version": 1,
		"binary": "webhook-deployer",
		"sha256": "abcd"
	}`)
	m, err := ParseManifest(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if m.ProviderType != "webhook-deployer" || m.ProtocolVersion != 1 {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

func TestParseManifest_InvalidJSON(t *testing.T) {
	if _, err := ParseManifest([]byte(`{not json`)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidate_MissingFields(t *testing.T) {
	m := &Manifest{ProviderType: "x"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for missing fields")
	}
}

func TestValidate_OSArchTogether(t *testing.T) {
	m := &Manifest{
		ProviderType:       "x",
		AccessProviderType: "x",
		Capability:         CapabilityDeploy,
		ProtocolVersion:    1,
		Binary:             "b",
		OS:                 "linux",
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error when only os specified")
	}
}

func TestParseManifest_WithNewFields(t *testing.T) {
	raw := []byte(`{
		"version": "0.1.0",
		"provider_type": "aliyun-cdn",
		"access_provider_type": "aliyun",
		"capability": "deploy",
		"display_name_key": "plugin.aliyun-cdn.name",
		"deploy_category": "cdn",
		"protocol_version": 1,
		"binary": "aliyun-cdn",
		"usages": ["dns", "hosting"],
		"priority": 10,
		"description": "Deploys certs to Aliyun CDN"
	}`)
	m, err := ParseManifest(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(m.Usages) != 2 || m.Usages[0] != "dns" || m.Usages[1] != "hosting" {
		t.Fatalf("unexpected usages: %v", m.Usages)
	}
	if m.Priority != 10 {
		t.Fatalf("unexpected priority: %d", m.Priority)
	}
	if m.Description != "Deploys certs to Aliyun CDN" {
		t.Fatalf("unexpected description: %q", m.Description)
	}
}

func TestParseManifest_WithoutNewFields(t *testing.T) {
	raw := []byte(`{
		"version": "0.1.0",
		"provider_type": "webhook-deployer",
		"access_provider_type": "webhook-deployer",
		"capability": "deploy",
		"display_name_key": "plugin.webhook-deployer.name",
		"deploy_category": "other",
		"protocol_version": 1,
		"binary": "webhook-deployer"
	}`)
	m, err := ParseManifest(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(m.Usages) != 0 {
		t.Fatalf("expected zero usages by default, got %v", m.Usages)
	}
	if m.Priority != 0 {
		t.Fatalf("expected zero priority by default, got %d", m.Priority)
	}
}

func TestValidate_InvalidUsages(t *testing.T) {
	m := &Manifest{
		ProviderType:       "x",
		AccessProviderType: "x",
		Capability:         CapabilityDeploy,
		ProtocolVersion:    1,
		Binary:             "b",
		Usages:             []string{"invalid"},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for invalid usages")
	}
}

func TestValidate_EmptyUsages(t *testing.T) {
	m := &Manifest{
		ProviderType:       "x",
		AccessProviderType: "x",
		Capability:         CapabilityDeploy,
		ProtocolVersion:    1,
		Binary:             "b",
		Usages:             []string{},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("empty usages should be valid: %v", err)
	}
}

func TestParseManifest_CapabilityForms(t *testing.T) {
	base := func(capability string) []byte {
		return []byte(`{
			"version": "0.1.0",
			"provider_type": "demo",
			"access_provider_type": "demo",
			"capability": ` + capability + `,
			"protocol_version": 3,
			"binary": "demo"
		}`)
	}

	m, err := ParseManifest(base(`"dns"`))
	if err != nil {
		t.Fatalf("parse string capability: %v", err)
	}
	if m.Capability != CapabilityDns {
		t.Fatalf("expected dns capability, got %q", m.Capability)
	}

	m, err = ParseManifest(base(`["notify"]`))
	if err != nil {
		t.Fatalf("parse single-element capability array: %v", err)
	}
	if m.Capability != CapabilityNotify {
		t.Fatalf("expected notify capability, got %q", m.Capability)
	}

	if _, err := ParseManifest(base(`["deploy", "dns"]`)); err == nil {
		t.Fatal("expected error for two capabilities")
	}

	if _, err := ParseManifest(base(`[]`)); err == nil {
		t.Fatal("expected error for empty capability array")
	}

	m, err = ParseManifest(base(`"bogus"`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for unknown capability")
	}

	m, err = ParseManifest([]byte(`{
		"version": "0.1.0",
		"provider_type": "demo",
		"access_provider_type": "demo",
		"protocol_version": 3,
		"binary": "demo"
	}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected validation error for missing capability")
	}
}
