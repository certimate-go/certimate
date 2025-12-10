package certifiers

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/certimate-go/certimate/internal/domain"
	dnscom "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/51dnscom"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderType51DNScom, func(options *ProviderFactoryOptions) (challenge.Provider, error) {
		credentials := domain.AccessConfigFor51DNScom{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := dnscom.NewChallenger(&dnscom.ChallengerConfig{
			ApiKey:                credentials.ApiKey,
			ApiSecret:             credentials.ApiSecret,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
