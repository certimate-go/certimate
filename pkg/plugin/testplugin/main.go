//go:build fakeplugin

package main

import (
	"context"
	"fmt"
	"os"

	githubplugin "github.com/hashicorp/go-plugin"

	"github.com/certimate-go/certimate/pkg/plugin"
)

type fakeDeployer struct {
	behavior string
}

func (f *fakeDeployer) GetMetadata(ctx context.Context) (*plugin.Metadata, error) {
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
	}, nil
}

func (f *fakeDeployer) GetConfigSchema(ctx context.Context) (*plugin.ConfigSchema, error) {
	return &plugin.ConfigSchema{
		AccessSchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"access"}`),
		DeploySchemaJSON: []byte(`{"schemaVersion":"form/v1","provider":"fake","category":"deploy"}`),
		I18n: map[string]map[string]string{
			"en": {"plugin.fake.name": "Fake Plugin"},
		},
	}, nil
}

func (f *fakeDeployer) Deploy(ctx context.Context, req *plugin.DeployRequest) (*plugin.DeployResult, error) {
	fmt.Fprintf(os.Stderr, "fakeplugin deploy behavior=%s access=%s\n", f.behavior, req.AccessConfigJSON)
	switch f.behavior {
	case "crash":
		fmt.Fprintln(os.Stderr, "fakeplugin crashing now")
		os.Exit(2)
	case "configerror":
		return nil, fmt.Errorf("bad config: missing field")
	}
	return &plugin.DeployResult{ExtendedDataJSON: `{"deployed":true}`}, nil
}

func main() {
	behavior := os.Getenv("FAKEPLUGIN_BEHAVIOR")
	if behavior == "" {
		behavior = "ok"
	}
	impl := &fakeDeployer{behavior: behavior}

	githubplugin.Serve(&githubplugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]githubplugin.Plugin{
			plugin.PluginName: &plugin.DeployerGRPCPlugin{Impl: impl},
		},
		GRPCServer: githubplugin.DefaultGRPCServer,
	})
}
