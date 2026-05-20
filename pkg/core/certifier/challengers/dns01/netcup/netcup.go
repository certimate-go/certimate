package netcup

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/netcup"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	CustomerNumber        string `json:"customerNumber"`
	ApiKey                string `json:"apiKey"`
	ApiPassword           string `json:"apiPassword"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := netcup.NewDefaultConfig()
	providerConfig.Customer = config.CustomerNumber
	providerConfig.Key = config.ApiKey
	providerConfig.Password = config.ApiPassword
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := netcup.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
