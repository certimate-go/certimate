package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	dplyimpl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-eomakers"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(
		domain.DeploymentProviderTypeTencentCloudEOMakers,
		func(options *ProviderFactoryOptions) (core.Deployer, error) {
			credentials := domain.AccessConfigForTencentCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			provider, err := dplyimpl.NewDeployer(&dplyimpl.DeployerConfig{
				SecretId:           credentials.SecretId,
				SecretKey:          credentials.SecretKey,
				ProjectId:          credentials.ProjectId,
				Endpoint:           xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
				MakersApiToken:     xmaps.GetString(options.ProviderExtendedConfig, "apiToken"),
				MakersProjectId:    xmaps.GetString(options.ProviderExtendedConfig, "projectId"),
				DomainMatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "domainMatchPattern"),
				Domains:            lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "domains"), ";"), func(s string, _ int) bool { return s != "" }),
				EnableMultipleSSL:  xmaps.GetBool(options.ProviderExtendedConfig, "enableMultipleSSL"),
			})
			return provider, err
		})
}
