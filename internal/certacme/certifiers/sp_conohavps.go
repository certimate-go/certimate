package certifiers

import (
	"fmt"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	conohavpsv2 "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/conohavpsv2"
	conohavpsv3 "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/conohavpsv3"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

func init() {
	ACMEDns01Registries.MustRegister(domain.ACMEDns01ProviderTypeConoHaVPS, func(options *ProviderFactoryOptions) (core.ACMEChallenger, error) {
		credentials := domain.AccessConfigForConoHaVPS{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		switch credentials.ApiVersion {
		case "2", "2.0", "v2", "v2.0":
			return conohavpsv2.NewChallenger(&conohavpsv2.ChallengerConfig{
				ApiUserName:           credentials.ApiUserName,
				ApiPassword:           credentials.ApiPassword,
				TenantId:              credentials.TenantId,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
		case "3", "3.0", "v3", "v3.0":
			return conohavpsv3.NewChallenger(&conohavpsv3.ChallengerConfig{
				ApiUserId:             credentials.ApiUserId,
				ApiUserName:           credentials.ApiUserName,
				ApiPassword:           credentials.ApiPassword,
				TenantId:              credentials.TenantId,
				TenantName:            credentials.TenantName,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
		default:
			return nil, fmt.Errorf("conohavps: unsupported api version: '%s'", credentials.ApiVersion)
		}
	})
}
