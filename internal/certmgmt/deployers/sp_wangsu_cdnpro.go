package deployers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	wangsucdnpro "github.com/certimate-go/certimate/pkg/core/deployer/providers/wangsu-cdnpro"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeWangsuCDNPro, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForWangsu{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := wangsucdnpro.NewDeployer(&wangsucdnpro.DeployerConfig{
			AccessKeyId:        credentials.AccessKeyId,
			AccessKeySecret:    credentials.AccessKeySecret,
			ApiKey:             credentials.ApiKey,
			Environment:        xmaps.GetOrDefaultString(options.ProviderExtendedConfig, "environment", "production"),
			DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
			Domain:             xmaps.GetString(options.ProviderExtendedConfig, "domain"),
			CertificateId:      xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
			WebhookId:          xmaps.GetString(options.ProviderExtendedConfig, "webhookId"),
		})
		return provider, err
	})
}
