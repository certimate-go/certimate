package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/email"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeEmail, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForEmail{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			SmtpHost:                 credentials.SmtpHost,
			SmtpPort:                 credentials.SmtpPort,
			SmtpTls:                  credentials.SmtpTls,
			Username:                 credentials.Username,
			Password:                 credentials.Password,
			SenderAddress:            credentials.SenderAddress,
			SenderName:               credentials.SenderName,
			ReceiverAddress:          xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "receiverAddress", credentials.ReceiverAddress),
			FileFormat:               xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "fileFormat", dplyimpl.FILE_FORMAT_PEM),
			PfxPassword:              xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "pfxPassword", ""),
			PfxEncoder:               xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "pfxEncoder", ""),
			JksAlias:                 xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "jksAlias", ""),
			JksKeypass:               xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "jksKeypass", ""),
			JksStorepass:             xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "jksStorepass", ""),
			AllowInsecureConnections: credentials.AllowInsecureConnections,
		})
		return provider, err
	})
}
