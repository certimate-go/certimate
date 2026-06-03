package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/ovhcloud"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeOVHcloud, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForOVHcloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			Endpoint:              credentials.Endpoint,
			AuthMethod:            credentials.AuthMethod,
			ApplicationKey:        credentials.ApplicationKey,
			ApplicationSecret:     credentials.ApplicationSecret,
			ConsumerKey:           credentials.ConsumerKey,
			ClientId:              credentials.ClientId,
			ClientSecret:          credentials.ClientSecret,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
