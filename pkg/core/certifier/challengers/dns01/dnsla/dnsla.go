package dnsla

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/dnsla/internal"
)

type ChallengerConfig struct {
	ApiId                 string `json:"apiId"`
	ApiSecret             string `json:"apiSecret"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.APIId = config.ApiId
	providerConfig.APISecret = config.ApiSecret
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
