package purgers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	prgimpl "github.com/certimate-go/certimate/pkg/core/purger/providers/tencentcloud-ssl"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.PurgeProviderTypeTencentCloudSSL, func(options *ProviderFactoryOptions) (core.Purger, error) {
		credentials := domain.AccessConfigForTencentCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := prgimpl.NewPurger(&prgimpl.PurgerConfig{
			SecretId:  credentials.SecretId,
			SecretKey: credentials.SecretKey,
			ProjectId: credentials.ProjectId,
			Endpoint:  xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
		})
		return provider, err
	})
}
