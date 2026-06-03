package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/namesilo"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeNameSilo, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForNameSilo{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			ApiKey:                credentials.ApiKey,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
