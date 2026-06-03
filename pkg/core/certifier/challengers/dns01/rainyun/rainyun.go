package rainyun

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/rainyun"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	ApiKey                string `json:"apiKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := rainyun.NewDefaultConfig()
	providerConfig.APIKey = config.ApiKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := rainyun.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
