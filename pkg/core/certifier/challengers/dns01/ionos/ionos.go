package ionos

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/ionos"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	ApiKeyPublicPrefix    string `json:"apiKeyPublicPrefix"`
	ApiKeySecret          string `json:"apiKeySecret"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := ionos.NewDefaultConfig()
	providerConfig.APIKey = fmt.Sprintf("%s.%s", config.ApiKeyPublicPrefix, config.ApiKeySecret)
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := ionos.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
