package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	akamaiedgedns "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/akamai-edgedns"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeAkamaiEdgeDNS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForAkamai{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := akamaiedgedns.NewChallenger(&akamaiedgedns.ChallengerConfig{
			Host:                  credentials.Host,
			ClientToken:           credentials.ClientToken,
			ClientSecret:          credentials.ClientSecret,
			AccessToken:           credentials.AccessToken,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})

	ACMEDns01Registries.MustRegisterAlias(domain.ACMEDns01ProviderTypeAkamai, domain.ACMEDns01ProviderTypeAkamaiEdgeDNS)
}
