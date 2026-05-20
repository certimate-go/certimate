package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/volcengine"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeVolcEngineDNS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForVolcEngine{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := volcengine.NewChallenger(&volcengine.ChallengerConfig{
			AccessKeyId:           credentials.AccessKeyId,
			SecretAccessKey:       credentials.SecretAccessKey,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})

	ACMEDns01Registries.MustRegisterAlias(domain.ACMEDns01ProviderTypeVolcEngine, domain.ACMEDns01ProviderTypeVolcEngineDNS)
}
