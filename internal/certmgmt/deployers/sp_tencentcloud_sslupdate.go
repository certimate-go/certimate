package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudsslupdate "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ssl-update"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.DeploymentProviderTypeTencentCloudSSLUpdate, func(options *ProviderFactoryOptions) (core.Deployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudsslupdate.NewDeployer(&tencentcloudsslupdate.DeployerConfig{
			SecretId:         credentials.SecretId,
			SecretKey:        credentials.SecretKey,
			Endpoint:         xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			CertificateId:    xmaps.GetString(options.ProviderExtendedConfig, "certificateId"),
			IsReplaced:       xmaps.GetBool(options.ProviderExtendedConfig, "isReplaced"),
			ResourceProducts: lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "resourceProducts"), ";"), func(s string, _ int) bool { return s != "" }),
			ResourceRegions:  lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "resourceRegions"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	})
}
