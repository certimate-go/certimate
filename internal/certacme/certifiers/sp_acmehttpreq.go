package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/acmehttpreq"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeACMEHttpReq, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForACMEHttpReq{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := acmehttpreq.NewChallenger(&acmehttpreq.ChallengerConfig{
			Endpoint:              credentials.Endpoint,
			Mode:                  credentials.Mode,
			Username:              credentials.Username,
			Password:              credentials.Password,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
		})
		return provider, err
	})
}
