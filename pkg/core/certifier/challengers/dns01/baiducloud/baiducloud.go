package baiducloud

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/baiducloud"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"secretAccessKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := baiducloud.NewDefaultConfig()
	providerConfig.AccessKeyID = config.AccessKeyId
	providerConfig.SecretAccessKey = config.SecretAccessKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := baiducloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
