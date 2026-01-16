package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/s3"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeS3, func(options *ProviderFactoryOptions) (deployer.Provider, error) {
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
			OutputFormat:                  xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "format", s3.OUTPUT_FORMAT_PEM),
			OutputCertObjectKey:           xmaps.GetString(options.ProviderExtendedConfig, "certObjectKey"),
			OutputServerCertObjectKey:     xmaps.GetString(options.ProviderExtendedConfig, "certObjectKeyForServerOnly"),
			OutputIntermediaCertObjectKey: xmaps.GetString(options.ProviderExtendedConfig, "certObjectKeyForIntermediaOnly"),
			OutputKeyObjectKey:            xmaps.GetString(options.ProviderExtendedConfig, "keyObjectKey"),
			PfxPassword:                   xmaps.GetString(options.ProviderExtendedConfig, "pfxPassword"),
			JksAlias:                      xmaps.GetString(options.ProviderExtendedConfig, "jksAlias"),
			JksKeypass:                    xmaps.GetString(options.ProviderExtendedConfig, "jksKeypass"),
			JksStorepass:                  xmaps.GetString(options.ProviderExtendedConfig, "jksStorepass"),
		})
		return provider, err
	})
}
