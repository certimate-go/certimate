package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/ionos"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeIONOS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForIONOS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			ApiKeyPublicPrefix:    credentials.ApiKeyPublicPrefix,
			ApiKeySecret:          credentials.ApiKeySecret,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
