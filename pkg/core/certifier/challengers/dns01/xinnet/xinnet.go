package xinnet

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/xinnet"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	AgentId               string `json:"agentId"`
	ApiPassword           string `json:"apiPassword"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := xinnet.NewDefaultConfig()
	providerConfig.AgentID = config.AgentId
	providerConfig.Secret = config.ApiPassword
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := xinnet.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
