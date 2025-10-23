package ionos

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/ionos"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengeProviderConfig struct {
	ApiKeyPublicPrefix    string `json:"apiKeyPublicPrefix"`
	ApiKeySecret          string `json:"apiKeySecret"`
	DnsPropagationTimeout int32  `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int32  `json:"dnsTTL,omitempty"`
}

func NewChallengeProvider(config *ChallengeProviderConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := ionos.NewDefaultConfig()
	providerConfig.APIKey = fmt.Sprintf("%s.%s", config.ApiKeyPublicPrefix, config.ApiKeySecret)
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = int(config.DnsTTL)
	}

	provider, err := ionos.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
