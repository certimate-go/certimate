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
			ShellEnv:                     xmaps.GetString(options.ProviderExtendedConfig, "shellEnv"),
			PreCommand:                   xmaps.GetString(options.ProviderExtendedConfig, "preCommand"),
			PostCommand:                  xmaps.GetString(options.ProviderExtendedConfig, "postCommand"),
			FileFormat:                   xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "fileFormat", local.FILE_FORMAT_PEM),
			FilePathForKey:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForKey"),
			FilePathForCrt:               xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrt"),
			FilePathForCrtOnlyServer:     xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyServer"),
			FilePathForCrtOnlyIntermedia: xmaps.GetString(options.ProviderExtendedConfig, "filePathForCrtOnlyIntermedia"),
			PfxPassword:                  xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			PfxEncoder:                   xmaps.GetString(options.ProviderExtendedConfig, "pfxEncoder"),
			JksAlias:                     xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:                   xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:                 xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
