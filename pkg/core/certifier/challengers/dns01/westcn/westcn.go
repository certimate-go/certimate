package westcn

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/westcn"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	Username              string `json:"username"`
	ApiPassword           string `json:"apiPassword"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := westcn.NewDefaultConfig()
	providerConfig.Username = config.Username
	providerConfig.Password = config.ApiPassword
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := westcn.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
