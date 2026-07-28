package pluginhost

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/plugin"
)

type pluginDeployerFactory struct {
	manager      *plugin.Manager
	dp           *plugin.DiscoveredPlugin
	providerType string
	logLevel     string
	logger       *slog.Logger
	mu           sync.Mutex
}

func (f *pluginDeployerFactory) new(options *deployers.ProviderFactoryOptions) (core.Deployer, error) {
	return &pluginDeployer{
		factory:  f,
		access:   options.ProviderAccessConfig,
		extended: options.ProviderExtendedConfig,
		logger:   f.logger,
	}, nil
}

type pluginDeployer struct {
	factory  *pluginDeployerFactory
	access   map[string]any
	extended map[string]any
	logger   *slog.Logger
}

func (d *pluginDeployer) SetLogger(logger *slog.Logger) {
	if logger != nil {
		d.logger = logger
	}
}

func (d *pluginDeployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*core.DeployerDeployResult, error) {
	d.factory.mu.Lock()
	defer d.factory.mu.Unlock()

	if d.logger != nil {
		d.logger.Info("deploying via plugin", slog.String("provider", d.factory.providerType))
	}

	req := &plugin.DeployRequest{
		LogLevel:           d.factory.logLevel,
		AccessConfigJSON:   marshalJSON(d.access),
		ExtendedConfigJSON: marshalJSON(d.extended),
		CertificatePEM:     certPEM,
		PrivateKeyPEM:      privkeyPEM,
	}

	res, err := d.factory.manager.Deploy(ctx, d.factory.dp, req, d.logger)
	if err != nil {
		return nil, err
	}

	result := &core.DeployerDeployResult{}
	if res != nil && res.ExtendedDataJSON != "" {
		var extended map[string]any
		if err := json.Unmarshal([]byte(res.ExtendedDataJSON), &extended); err == nil {
			result.ExtendedData = extended
		}
	}
	return result, nil
}

func marshalJSON(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}
