package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/s3"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeS3, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForS3{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := s3.NewDeployer(&s3.DeployerConfig{
			Endpoint:                      credentials.Endpoint,
			AccessKey:                     credentials.AccessKey,
			SecretKey:                     credentials.SecretKey,
			SignatureVersion:              credentials.SignatureVersion,
			UsePathStyle:                  credentials.UsePathStyle,
			AllowInsecureConnections:      credentials.AllowInsecureConnections,
			Region:                        xmaps.GetString(options.ProviderExtendedConfig, "region"),
			Bucket:                        xmaps.GetString(options.ProviderExtendedConfig, "bucket"),
			FileFormat:                    xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "fileFormat", s3.FILE_FORMAT_PEM),
			ObjectKeyForKey:               xmaps.GetString(options.ProviderExtendedConfig, "objectKeyForKey"),
			ObjectKeyForCrt:               xmaps.GetString(options.ProviderExtendedConfig, "objectKeyForCrt"),
			ObjectKeyForCrtOnlyServer:     xmaps.GetString(options.ProviderExtendedConfig, "objectKeyForCrtOnlyServer"),
			ObjectKeyForCrtOnlyIntermedia: xmaps.GetString(options.ProviderExtendedConfig, "objectKeyForCrtOnlyIntermedia"),
			PfxPassword:                   xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                      xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:                    xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:                  xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
