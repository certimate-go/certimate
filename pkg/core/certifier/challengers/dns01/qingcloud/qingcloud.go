package qingcloud

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/qingcloud/internal"
)

type ChallengerConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"apiPassword"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.AccessKey = config.AccessKeyId
	providerConfig.AccessSecret = config.SecretAccessKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := internal.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
