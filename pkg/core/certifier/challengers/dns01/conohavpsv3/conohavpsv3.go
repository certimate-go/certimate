package conohavpsv3

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/conohavpsv3/internal"
)

type ChallengerConfig struct {
	ApiUserId             string `json:"apiUserId"`
	ApiUserName           string `json:"apiUserName"`
	ApiPassword           string `json:"apiPassword"`
	TenantId              string `json:"tenantId"`
	TenantName            string `json:"tenantName"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.UserID = config.ApiUserId
	providerConfig.UserName = config.ApiUserName
	providerConfig.Password = config.ApiPassword
	providerConfig.TenantID = config.TenantId
	providerConfig.TenantName = config.TenantName
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
