package applicators

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/certimate-go/certimate/internal/domain"
	akamaiedgedns "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/akamai-edgedns"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	if err := ACMEDns01Registries.Register(domain.ACMEDns01ProviderTypeAkamaiEdgeDNS, func(options *ProviderFactoryOptions) (challenge.Provider, error) {
		credentials := domain.AccessConfigForAkamai{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := akamaiedgedns.NewChallengeProvider(&akamaiedgedns.ChallengeProviderConfig{
			Host:                  credentials.Host,
			ClientToken:           credentials.ClientToken,
			ClientSecret:          credentials.ClientSecret,
			AccessToken:           credentials.AccessToken,
			DnsPropagationTimeout: int(options.DnsPropagationTimeout),
			DnsTTL:                int(options.DnsTTL),
		})
		return provider, err
	}); err != nil {
		panic(err)
	}

	if err := ACMEDns01Registries.RegisterAlias(domain.ACMEDns01ProviderTypeAkamai, domain.ACMEDns01ProviderTypeAkamaiEdgeDNS); err != nil {
		panic(err)
	}
}
