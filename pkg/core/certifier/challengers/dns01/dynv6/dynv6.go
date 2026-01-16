package dynv6

import (
	"errors"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/dynv6/internal"
)

type ChallengerConfig struct {
	HttpToken             string `json:"httpToken"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.HTTPToken = config.HttpToken
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := internal.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
