package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudupathx "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-upathx"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeUCloudUPathX, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ucloudupathx.NewDeployer(&ucloudupathx.DeployerConfig{
			PrivateKey:    credentials.PrivateKey,
			PublicKey:     credentials.PublicKey,
			ProjectId:     credentials.ProjectId,
			AcceleratorId: xmaps.GetString(options.ProviderExtendedConfig, "acceleratorId"),
			ListenerPort:  xmaps.GetInt32(options.ProviderExtendedConfig, "listenerPort"),
		})
		return provider, err
	})
}
