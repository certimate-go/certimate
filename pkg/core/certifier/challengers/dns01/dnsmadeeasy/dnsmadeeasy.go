package dnsmadeeasy

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/dnsmadeeasy"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	ApiKey                string `json:"apiKey"`
	ApiSecret             string `json:"apiSecret"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := dnsmadeeasy.NewDefaultConfig()
	providerConfig.APIKey = config.ApiKey
	providerConfig.APISecret = config.ApiSecret
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := dnsmadeeasy.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
