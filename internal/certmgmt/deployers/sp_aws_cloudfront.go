package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	awscloudfront "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-cloudfront"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAWSCloudFront, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAWS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := awscloudfront.NewDeployer(&awscloudfront.DeployerConfig{
			AccessKeyId:       credentials.AccessKeyId,
			SecretAccessKey:   credentials.SecretAccessKey,
			Region:            xmaps.GetString(options.ProviderExtendedConfig, "region"),
			DistributionId:    xmaps.GetString(options.ProviderExtendedConfig, "distributionId"),
			CertificateSource: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "certificateSource", "ACM"),
		})
		return provider, err
	})
}
