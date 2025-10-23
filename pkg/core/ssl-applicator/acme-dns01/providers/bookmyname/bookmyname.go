package bookmyname

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/bookmyname"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengeProviderConfig struct {
	Username              string `json:"username"`
	Password              string `json:"password"`
	DnsPropagationTimeout int32  `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int32  `json:"dnsTTL,omitempty"`
}

func NewChallengeProvider(config *ChallengeProviderConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := bookmyname.NewDefaultConfig()
	providerConfig.Username = config.Username
	providerConfig.Password = config.Password
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = int(config.DnsTTL)
	}

	provider, err := bookmyname.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
