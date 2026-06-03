package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	chlgimpl "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/cmcccloud"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeCMCCCloudDNS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForCMCCCloud{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := chlgimpl.NewChallenger(&chlgimpl.ChallengerConfig{
			AccessKeyId:           credentials.AccessKeyId,
			AccessKeySecret:       credentials.AccessKeySecret,
			DnsPropagationTimeout: options.DnsPropagationTimeout,
			DnsTTL:                options.DnsTTL,
		})
		return provider, err
	})

	ACMEDns01Registries.MustRegisterAlias(domain.ACMEDns01ProviderTypeCMCCCloud, domain.ACMEDns01ProviderTypeCMCCCloudDNS)
}
