//go:build fakeplugin

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	githubplugin "github.com/hashicorp/go-plugin"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func statePath() string {
	return envOr("FAKEPLUGIN_STATE_FILE", filepath.Join(os.TempDir(), "fakeplugin-state.json"))
}

func appendState(event string) {
	f, err := os.OpenFile(statePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n", event)
}

func metadata() *plugin.Metadata {
	accessType := os.Getenv("FAKEPLUGIN_ACCESS_TYPE")
	if accessType == "" {
		accessType = os.Getenv("FAKEPLUGIN_PROVIDER_TYPE")
	}
	return &plugin.Metadata{
		ProviderType:         os.Getenv("FAKEPLUGIN_PROVIDER_TYPE"),
		AccessProviderType:   accessType,
		ProtocolVersion:      plugin.ProtocolVersion,
		DeployCategory:       "other",
		DeployDisplayNameKey: "plugin.fake.name",
		AccessDisplayNameKey: "plugin.fake.name",
	}
}

func omitAccessSchema() bool {
	return os.Getenv("FAKEPLUGIN_OMIT_ACCESS_SCHEMA") == "1"
}

func configSchema() *plugin.ConfigSchema {
	var accessSchema []byte
	if !omitAccessSchema() {
		accessSchema = []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"access","schema":{"columns":[{"name":"apiKey","valueType":"string","required":true,"secret":true}]}}`)
	}
	return &plugin.ConfigSchema{
		AccessSchemaJSON: accessSchema,
		DeploySchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"deploy","schema":{"columns":[{"name":"endpoint","valueType":"string","required":true}]}}`),
		I18n: map[string]map[string]string{
			"en": {"plugin.fake.name": "Fake Plugin"},
		},
	}
}

type fakeDeployer struct {
	behavior string
}

func (f *fakeDeployer) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return metadata(), nil
}

func (f *fakeDeployer) GetConfigSchema(ctx context.Context) (*plugin.ConfigSchema, error) {
	return configSchema(), nil
}

func (f *fakeDeployer) Deploy(ctx context.Context, req *plugin.DeployRequest, logger *slog.Logger) (*plugin.DeployResult, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin deploy behavior=%s access=%s\n", f.behavior, req.AccessConfigJSON)
	if logger != nil {
		logger.Info("fakeplugin deploy starting", slog.String("behavior", f.behavior))
	}
	switch f.behavior {
	case "crash":
		fmt.Fprintln(os.Stderr, "fakeplugin crashing now")
		os.Exit(2)
	case "configerror":
		return nil, fmt.Errorf("bad config: missing field")
	case "error-echo":
		return nil, fmt.Errorf("remote rejected access %s and key %s", req.AccessConfigJSON, req.PrivateKeyPEM)
	}
	appendState("deploy")
	return &plugin.DeployResult{ExtendedDataJSON: `{"deployed":true}`}, nil
}

type fakeDnsProvider struct {
	behavior string
}

func (f *fakeDnsProvider) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return metadata(), nil
}

func (f *fakeDnsProvider) GetConfigSchema(ctx context.Context) (*plugin.ConfigSchema, error) {
	return configSchema(), nil
}

func (f *fakeDnsProvider) Present(ctx context.Context, req *plugin.DnsRequest) error {
	fmt.Fprintf(os.Stderr, "fakeplugin present pid=%d domain=%s\n", os.Getpid(), req.Domain)
	switch f.behavior {
	case "present-crash":
		appendState("present")
		fmt.Fprintln(os.Stderr, "fakeplugin exiting shortly after present")
		time.AfterFunc(150*time.Millisecond, func() { os.Exit(2) })
		return nil
	case "present-error":
		if req.Domain == "failed.example.test" {
			appendState("present:error")
			return fmt.Errorf("fake present failure")
		}
	}
	appendState("present")
	return nil
}

func (f *fakeDnsProvider) CleanUp(ctx context.Context, req *plugin.DnsRequest) error {
	fmt.Fprintf(os.Stderr, "fakeplugin cleanup pid=%d domain=%s\n", os.Getpid(), req.Domain)
	appendState("cleanup")
	return nil
}

type fakeNotifier struct {
	behavior string
}

func (f *fakeNotifier) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
	return metadata(), nil
}

func (f *fakeNotifier) GetConfigSchema(ctx context.Context) (*plugin.ConfigSchema, error) {
	return configSchema(), nil
}

func (f *fakeNotifier) Notify(ctx context.Context, req *plugin.NotifyRequest) (*plugin.NotifyResult, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin notify subject=%s\n", req.Subject)
	switch f.behavior {
	case "notify-error":
		return nil, fmt.Errorf("fake notify failure")
	}
	appendState("notify:" + req.Subject + "|" + req.Message)
	return &plugin.NotifyResult{}, nil
}

func main() {
	behavior := envOr("FAKEPLUGIN_BEHAVIOR", "ok")
	capability := envOr("FAKEPLUGIN_CAPABILITY", plugin.CapabilityDeploy)

	var served githubplugin.Plugin
	switch capability {
	case plugin.CapabilityDns:
		served = &plugin.DnsGRPCPlugin{Impl: &fakeDnsProvider{behavior: behavior}}
	case plugin.CapabilityNotify:
		served = &plugin.NotifyGRPCPlugin{Impl: &fakeNotifier{behavior: behavior}}
	default:
		served = &plugin.DeployerGRPCPlugin{Impl: &fakeDeployer{behavior: behavior}}
	}

	githubplugin.Serve(&githubplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]githubplugin.Plugin{
			plugin.ServiceNameForCapability(capability): served,
		},
		GRPCServer: githubplugin.DefaultGRPCServer,
	})
}
