package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/goedge"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeGoEdge, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForGoEdge{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiRole:                  credentials.ApiRole,
			AccessKeyId:              credentials.AccessKeyId,
			AccessKey:                credentials.AccessKey,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DeployTarget:             xmaps.GetString(options.ProviderExtendedConfig, "deployTarget"),
			CertificateId:            xmaps.GetInt64(options.ProviderExtendedConfig, "certificateId"),
		})
		return provider, err
	})
}
