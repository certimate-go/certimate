package gandinet

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/gandiv5"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	PersonalAccessToken   string `json:"personalAccessToken"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := gandiv5.NewDefaultConfig()
	providerConfig.PersonalAccessToken = config.PersonalAccessToken
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := gandiv5.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
