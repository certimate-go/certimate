package plugin

import (
	"testing"
)

func TestParseManifest_Valid(t *testing.T) {
	raw := []byte(`{
		"version": "0.1.0",
		"provider_type": "webhook-deployer",
		"access_provider_type": "webhook-deployer",
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
		ProtocolVersion:    1,
		Binary:             "b",
		OS:                 "linux",
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error when only os specified")
	}
}
