package deployers

import (
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/local"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeLocal, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		provider, err := local.NewDeployer(&local.DeployerConfig{
			ShellEnv:                 xmaps.GetString(options.ProviderExtendedConfig, "shellEnv"),
			PreCommand:               xmaps.GetString(options.ProviderExtendedConfig, "preCommand"),
			PostCommand:              xmaps.GetString(options.ProviderExtendedConfig, "postCommand"),
			OutputFormat:             xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", local.OUTPUT_FORMAT_PEM),
			OutputCertPath:           xmaps.GetString(options.ProviderExtendedConfig, "certPath"),
			OutputServerCertPath:     xmaps.GetString(options.ProviderExtendedConfig, "certPathForServerOnly"),
			OutputIntermediaCertPath: xmaps.GetString(options.ProviderExtendedConfig, "certPathForIntermediaOnly"),
			OutputKeyPath:            xmaps.GetString(options.ProviderExtendedConfig, "keyPath"),
			PfxPassword:              xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                 xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:               xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:             xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
