package purgers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	pgrimpl "github.com/certimate-go/certimate/pkg/core/purger/providers/huaweicloud-waf"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.PurgeProviderTypeHuaweiCloudWAF, func(options *ProviderFactoryOptions) (core.Purger, error) {
		credentials := domain.AccessConfigForHuaweiCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := pgrimpl.NewPurger(&pgrimpl.PurgerConfig{
			AccessKeyId:         credentials.AccessKeyId,
			SecretAccessKey:     credentials.SecretAccessKey,
			EnterpriseProjectId: credentials.EnterpriseProjectId,
			Region:              xmaps.GetString(options.ProviderExtendedConfig, "region"),
		})
		return provider, err
	})
}
