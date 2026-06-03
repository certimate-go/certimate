package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/technitiumdns"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeTechnitiumDNS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForTechnitiumDNS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			ServerUrl:                credentials.ServerUrl,
			ApiToken:                 credentials.ApiToken,
			AllowInsecureConnections: credentials.AllowInsecureConnections,
			DnsPropagationTimeout:    options.DnsPropagationTimeout,
			DnsTTL:                   options.DnsTTL,
		})
		return provider, err
	})
}
