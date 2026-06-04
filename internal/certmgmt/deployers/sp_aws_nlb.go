package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-nlb"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAWSNLB, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAWS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:       credentials.AccessKeyId,
			SecretAccessKey:   credentials.SecretAccessKey,
			Region:            xmaps.GetString(options.ProviderExtendedConfig, "region"),
			LoadbalancerArn:   xmaps.GetString(options.ProviderExtendedConfig, "loadbalancerArn"),
			ListenerArn:       xmaps.GetString(options.ProviderExtendedConfig, "listenerArn"),
			CertificateSource: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "certificateSource", dplyimpl.CERTIFICATE_SOURCE_ACM),
			IsDefault:         xmaps.GetBool(options.ProviderExtendedConfig, "isDefault"),
		})
		return provider, err
	})
}
