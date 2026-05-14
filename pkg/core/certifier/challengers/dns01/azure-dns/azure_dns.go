package azuredns

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/azuredns"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	azenv "github.com/certimate-go/certimate/pkg/sdk3rd/azure/env"
)

type ChallengerConfig struct {
	TenantId              string `json:"tenantId"`
	ClientId              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret"`
	SubscriptionId        string `json:"subscriptionId,omitempty"`
	ResourceGroupName     string `json:"resourceGroupName,omitempty"`
	CloudName             string `json:"cloudName,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := azuredns.NewDefaultConfig()
	providerConfig.AuthMethod = "env"
	providerConfig.TenantID = config.TenantId
	providerConfig.ClientID = config.ClientId
	providerConfig.ClientSecret = config.ClientSecret
	providerConfig.SubscriptionID = config.SubscriptionId
	providerConfig.ResourceGroup = config.ResourceGroupName
	if config.CloudName != "" {
		env, err := azenv.GetCloudEnvConfiguration(config.CloudName)
		if err != nil {
			return nil, err
		}
		providerConfig.Environment = env
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := azuredns.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
