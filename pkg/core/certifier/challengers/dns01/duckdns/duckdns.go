package namedotcom

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/duckdns"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	Token                 string `json:"token"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := duckdns.NewDefaultConfig()
	providerConfig.Token = config.Token
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}

	provider, err := duckdns.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
