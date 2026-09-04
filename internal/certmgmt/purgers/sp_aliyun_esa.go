package purgers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	prgimpl "github.com/certimate-go/certimate/pkg/core/purger/providers/aliyun-esa"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.PurgeProviderTypeAliyunESA, func(options *ProviderFactoryOptions) (core.Purger, error) {
		credentials := domain.AccessConfigForAliyun{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := prgimpl.NewPurger(&prgimpl.PurgerConfig{
			AccessKeyId:     credentials.AccessKeyId,
			AccessKeySecret: credentials.AccessKeySecret,
			Region:          xmaps.GetString(options.ProviderExtendedConfig, "region"),
			SiteId:          xmaps.GetInt64(options.ProviderExtendedConfig, "siteId"),
		})
		return provider, err
	})
}
