package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/synologydsm"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeSynologyDSM, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForSynologyDSM{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := synologydsm.NewDeployer(&synologydsm.DeployerConfig{
			ServerUrl:                  credentials.ServerUrl,
			Username:                   credentials.Username,
			Password:                   credentials.Password,
			TotpSecret:                 credentials.TotpSecret,
			AllowInsecureConnections:   credentials.AllowInsecureConnections,
			CertificateIdOrDescription: xmaps.GetString(options.ProviderExtendedConfig, "certificateIdOrDesc"),
			IsDefault:                  xmaps.GetBool(options.ProviderExtendedConfig, "isDefault"),
		})
		return provider, err
	})
}
