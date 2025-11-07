package applicators

import (
	"fmt"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/edgedns"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
	"github.com/go-acme/lego/v4/challenge"
	"time"
)

func init() {
	if err := ACMEDns01Registries.Register(domain.ACMEDns01ProviderTypeAkamai, func(options *ProviderFactoryOptions) (challenge.Provider, error) {
		credentials := domain.AccessConfigForAkamai{}
		if err := xmaps.Populate(options.ProviderAccessConfig, &credentials); err != nil {
			return nil, fmt.Errorf("failed to populate provider access config: %w", err)
		}

		provider, err := edgedns.NewChallengeProvider(&edgedns.ChallengeProviderConfig{
			ClientToken:           credentials.ClientToken,
			ClientSecret:          credentials.ClientSecret,
			AccessToken:           credentials.AccessToken,
			Host:                  credentials.Host,
			DnsPropagationTimeout: time.Duration(options.DnsPropagationTimeout),
			DnsTTL:                int(options.DnsTTL),
		})
		return provider, err
	}); err != nil {
		panic(err)
	}
}
