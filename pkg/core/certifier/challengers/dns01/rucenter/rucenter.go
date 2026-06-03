package rucenter

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/nicru"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	Username              string `json:"username"`
	Password              string `json:"password"`
	ApplicationId         string `json:"applicationId"`
	ApplicationToken      string `json:"applicationToken"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := nicru.NewDefaultConfig()
	providerConfig.Username = config.Username
	providerConfig.Password = config.Password
	providerConfig.ServiceID = config.ApplicationId
	providerConfig.Secret = config.ApplicationToken
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := nicru.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
