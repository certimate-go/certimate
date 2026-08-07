package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/rfc2136"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeRFC2136, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForRFC2136{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			Host:                  credentials.Host,
			Port:                  credentials.Port,
			TsigAlgorithm:         credentials.TsigAlgorithm,
			TsigKey:               credentials.TsigKey,
			TsigSecret:            credentials.TsigSecret,
			TsigGssRealm:          credentials.TsigGssRealm,
			TsigGssUsername:       credentials.TsigGssUsername,
			TsigGssPassword:       credentials.TsigGssPassword,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})
}
