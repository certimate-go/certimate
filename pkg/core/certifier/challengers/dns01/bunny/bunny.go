package bunny

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/bunny"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	ApiKey                string `json:"apiKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := bunny.NewDefaultConfig()
	providerConfig.APIKey = config.ApiKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := bunny.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
