package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-iam"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeAWSIAM, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForAWS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			SecretAccessKey: credentials.SecretAccessKey,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			CertificatePath: xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "certificatePath", "/"),
		})
		return provider, err
	})
}
