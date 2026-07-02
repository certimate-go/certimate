package baishancdn

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestFilterWildcardDomainForBaishanAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		domain string
		want   string
	}{
		{
			name:   "wildcard domain",
			domain: "*.example.com",
			want:   ".example.com",
		},
		{
			name:   "already compliant wildcard domain",
			domain: ".example.com",
			want:   ".example.com",
		},
		{
			name:   "normal domain",
			domain: "example.com",
			want:   "example.com",
		},
		{
			name:   "non wildcard prefix",
			domain: "*example.com",
			want:   "*example.com",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := filterWildcardDomainForBaishanAPI(tt.domain)
			if got != tt.want {
				t.Fatalf("filterWildcardDomainForBaishanAPI(%q) = %q, want %q", tt.domain, got, tt.want)
			}
		})
	}
}

func TestGetFilteredDomainForAPI(t *testing.T) {
	t.Parallel()

	t.Run("logs filtered wildcard domain", func(t *testing.T) {
		var logbuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logbuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		deployer, err := NewDeployer(&DeployerConfig{
			ApiToken: "test-token",
			Domain:   "*.example.com",
		})
		if err != nil {
			t.Fatalf("NewDeployer() error = %v", err)
		}
		deployer.SetLogger(logger)

		got := deployer.getFilteredDomainForAPI("cdn.GetDomainConfig", deployer.config.Domain)
		if got != ".example.com" {
			t.Fatalf("getFilteredDomainForAPI() = %q, want %q", got, ".example.com")
		}

		var record map[string]any
		if err := json.Unmarshal(logbuf.Bytes(), &record); err != nil {
			t.Fatalf("failed to unmarshal log record: %v", err)
		}

		if record["level"] != "INFO" {
			t.Fatalf("log level = %v, want %q", record["level"], "INFO")
		}
		if record["msg"] != "baishan cdn wildcard domain prefix filtered" {
			t.Fatalf("log msg = %v, want %q", record["msg"], "baishan cdn wildcard domain prefix filtered")
		}
		if record["apiIdentifier"] != "cdn.GetDomainConfig" {
			t.Fatalf("log apiIdentifier = %v, want %q", record["apiIdentifier"], "cdn.GetDomainConfig")
		}
		if record["originalDomain"] != "*.example.com" {
			t.Fatalf("log originalDomain = %v, want %q", record["originalDomain"], "*.example.com")
		}
		if record["filteredDomain"] != ".example.com" {
			t.Fatalf("log filteredDomain = %v, want %q", record["filteredDomain"], ".example.com")
		}
		if _, ok := record["filterTime"]; !ok {
			t.Fatal("log missing filterTime field")
		}
	})

	t.Run("does not log for normal domain", func(t *testing.T) {
		var logbuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logbuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		deployer, err := NewDeployer(&DeployerConfig{
			ApiToken: "test-token",
			Domain:   "example.com",
		})
		if err != nil {
			t.Fatalf("NewDeployer() error = %v", err)
		}
		deployer.SetLogger(logger)

		got := deployer.getFilteredDomainForAPI("cdn.GetDomainConfig", deployer.config.Domain)
		if got != "example.com" {
			t.Fatalf("getFilteredDomainForAPI() = %q, want %q", got, "example.com")
		}
		if logbuf.Len() != 0 {
			t.Fatalf("expected no log output, got %s", logbuf.String())
		}
	})
}
