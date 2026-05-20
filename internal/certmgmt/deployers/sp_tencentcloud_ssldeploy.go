package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudssldeploy "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ssl-deploy"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudSSLDeploy, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudssldeploy.NewDeployer(&tencentcloudssldeploy.DeployerConfig{
			SecretId:        credentials.SecretId,
			SecretKey:       credentials.SecretKey,
			Endpoint:        xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			ResourceProduct: xmaps.GetString(options.ProviderExtendedConfig, "resourceProduct"),
			ResourceIds:     lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "resourceIds"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	})
}
