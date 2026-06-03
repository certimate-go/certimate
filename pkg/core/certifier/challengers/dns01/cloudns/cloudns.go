package cloudns

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/cloudns"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	AuthId                string `json:"authId"`
	AuthPassword          string `json:"authPassword"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := cloudns.NewDefaultConfig()
	providerConfig.AuthID = config.AuthId
	providerConfig.AuthPassword = config.AuthPassword
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := cloudns.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
