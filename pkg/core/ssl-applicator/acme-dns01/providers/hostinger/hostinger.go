package hostinger

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/hostinger"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengeProviderConfig struct {
	ApiToken              string `json:"apiToken"`
	DnsPropagationTimeout int32  `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int32  `json:"dnsTTL,omitempty"`
}

func NewChallengeProvider(config *ChallengeProviderConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := hostinger.NewDefaultConfig()
	providerConfig.APIToken = config.ApiToken
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = int(config.DnsTTL)
	}

	provider, err := hostinger.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
