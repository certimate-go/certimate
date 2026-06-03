package gname

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/gname"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	AppId                 string `json:"appId"`
	AppKey                string `json:"appKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := gname.NewDefaultConfig()
	providerConfig.AppID = config.AppId
	providerConfig.AppKey = config.AppKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := gname.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
