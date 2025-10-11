package deployers

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	tencentcloudeo "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-eo"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	if err := Registries.Register(domain.DeploymentProviderTypeTencentCloudEO, func(options *ProviderFactoryOptions) (core.SSLDeployer, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := tencentcloudeo.NewSSLDeployerProvider(&tencentcloudeo.SSLDeployerProviderConfig{
			SecretId:     credentials.SecretId,
			SecretKey:    credentials.SecretKey,
			Endpoint:     xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
			ZoneId:       xmaps.GetString(options.ProviderExtendedConfig, "zoneId"),
			MatchPattern: xmaps.GetString(options.ProviderExtendedConfig, "matchPattern"),
			Domains:      lo.Filter(strings.Split(xmaps.GetString(options.ProviderExtendedConfig, "domains"), ";"), func(s string, _ int) bool { return s != "" }),
		})
		return provider, err
	}); err != nil {
		panic(err)
	}
}
