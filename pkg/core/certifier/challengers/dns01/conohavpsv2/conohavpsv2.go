package conohavpsv2

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/conoha"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	ApiUserName           string `json:"apiUserName"`
	ApiPassword           string `json:"apiPassword"`
	TenantId              string `json:"tenantId"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := conoha.NewDefaultConfig()
	providerConfig.Username = config.ApiUserName
	providerConfig.Password = config.ApiPassword
	providerConfig.TenantID = config.TenantId
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := conoha.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
