package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/ftp"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeFTP, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
		credentials := domain.AccessConfigForFTP{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ftp.NewDeployer(&ftp.DeployerConfig{
			FtpHost:                  credentials.Host,
			FtpPort:                  credentials.Port,
			FtpUsername:              credentials.Username,
			FtpPassword:              credentials.Password,
			OutputFormat:             xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", ftp.OUTPUT_FORMAT_PEM),
			OutputKeyPath:            xmaps.GetString(options.ProviderExtendedConfig, "keyPath"),
			OutputCertPath:           xmaps.GetString(options.ProviderExtendedConfig, "certPath"),
			OutputServerCertPath:     xmaps.GetString(options.ProviderExtendedConfig, "certPathForServerOnly"),
			OutputIntermediaCertPath: xmaps.GetString(options.ProviderExtendedConfig, "certPathForIntermediaOnly"),
			PfxPassword:              xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                 xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:               xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:             xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
