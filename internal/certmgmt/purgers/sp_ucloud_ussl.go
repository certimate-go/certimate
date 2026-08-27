package purgers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	pgrimpl "github.com/certimate-go/certimate/pkg/core/purger/providers/ucloud-ussl"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	Registries.MustRegister(domain.PurgeProviderTypeUCloudUSSL, func(options *ProviderFactoryOptions) (core.Purger, error) {
		credentials := domain.AccessConfigForUCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := pgrimpl.NewPurger(&pgrimpl.PurgerConfig{
			PrivateKey: credentials.PrivateKey,
			PublicKey:  credentials.PublicKey,
			ProjectId:  credentials.ProjectId,
			Endpoint:   xmaps.GetString(options.ProviderExtendedConfig, "endpoint"),
		})
		return provider, err
	})
}
