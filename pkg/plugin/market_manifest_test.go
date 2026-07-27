package plugin

import (
	"testing"
)

func TestParseMarketManifest(t *testing.T) {
	data := []byte(`{
		"version": "1.0.0",
		"provider_type": "webhook-deployer",
		"access_provider_type": "webhook",
		"display_name_key": "plugin.webhook-deployer.name",
		"deploy_category": "other",
		"protocol_version": 1,
		"binary": "webhook-deployer",
		"description": "Deploy via webhook",
		"release": {
			"repo": "certimate-go/plugins",
			"tag": "webhook-deployer/v1.0.0",
			"assets": {
				"linux/amd64": "webhook-deployer_linux_amd64",
				"darwin/arm64": "webhook-deployer_darwin_arm64"
			}
		}
	}`)

	mm, err := ParseMarketManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mm.ProviderType != "webhook-deployer" {
		t.Errorf("expected provider_type webhook-deployer, got %s", mm.ProviderType)
	}
	if mm.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", mm.Version)
	}
	if mm.Release == nil {
		t.Fatal("expected non-nil release")
	}
	if mm.Release.Repo != "certimate-go/plugins" {
		t.Errorf("expected repo certimate-go/plugins, got %s", mm.Release.Repo)
	}
	if mm.Release.Tag != "webhook-deployer/v1.0.0" {
		t.Errorf("expected tag webhook-deployer/v1.0.0, got %s", mm.Release.Tag)
	}
	if len(mm.Release.Assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(mm.Release.Assets))
	}
}

func TestParseMarketManifest_NoRelease(t *testing.T) {
	data := []byte(`{"version": "1.0.0", "provider_type": "test"}`)
	_, err := ParseMarketManifest(data)
	if err == nil {
		t.Fatal("expected error for manifest without release block")
	}
}

func TestParseMarketManifest_InvalidJSON(t *testing.T) {
	_, err := ParseMarketManifest([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateProviderType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "webhook-deployer", true},
		{"with dots", "com.example.plugin", true},
		{"with underscores", "my_plugin_v2", true},
		{"alphanumeric only", "plugin123", true},
		{"empty", "", false},
		{"path traversal", "../../etc/passwd", false},
		{"with slash", "foo/bar", false},
		{"starts with dot", ".hidden", false},
		{"with spaces", "my plugin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderType(tt.input)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestValidateReleaseRepo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid", "certimate-go/plugins", true},
		{"valid with dots", "certimate-go/my.plugin-repo_v2", true},
		{"empty", "", false},
		{"wrong org", "other-org/plugins", false},
		{"path traversal", "../etc/passwd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReleaseRepo(tt.input)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestAssetKey(t *testing.T) {
	key := AssetKey("linux", "amd64")
	if key != "linux/amd64" {
		t.Errorf("expected linux/amd64, got %s", key)
	}
}

func TestValidateBinaryName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "webhook-deployer", true},
		{"with arch suffix", "webhook-deployer_linux_amd64", true},
		{"with exe", "webhook-deployer.exe", true},
		{"empty", "", false},
		{"path traversal", "../evil", false},
		{"absolute path", "/bin/sh", false},
		{"with slash", "foo/bar", false},
		{"parent dir", "..", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBinaryName(tt.input)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestParseMarketManifest_WithChecksums(t *testing.T) {
	data := []byte(`{
		"version": "1.0.0",
		"provider_type": "webhook-deployer",
		"access_provider_type": "webhook",
		"protocol_version": 1,
		"binary": "webhook-deployer",
		"release": {
			"repo": "certimate-go/plugins",
			"tag": "v1.0.0",
			"assets": {"linux/amd64": "webhook-deployer_linux_amd64"},
			"checksums": {"linux/amd64": "abc123"}
		}
	}`)

	mm, err := ParseMarketManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := mm.Release.Checksums["linux/amd64"]; got != "abc123" {
		t.Errorf("expected checksum abc123, got %s", got)
	}
}
