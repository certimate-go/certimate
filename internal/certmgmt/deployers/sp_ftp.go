package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/ftp"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeFTP, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForFTP{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := ftp.NewDeployer(&ftp.DeployerConfig{
			FtpHost:                      credentials.Host,
			FtpPort:                      credentials.Port,
			FtpUsername:                  credentials.Username,
			FtpPassword:                  credentials.Password,
			FileFormat:                   xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "fileFormat", ftp.FILE_FORMAT_PEM),
			FilePathForKey:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForKey"),
			FilePathForCrt:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrt"),
			FilePathForCrtOnlyServer:     xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyServer"),
			FilePathForCrtOnlyIntermedia: xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyIntermedia"),
			PfxPassword:                  xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                     xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:                   xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:                 xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
