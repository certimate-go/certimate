package azure

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/azuredns"

	"github.com/certimate-go/certimate/pkg/core"
	xazure "github.com/certimate-go/certimate/pkg/utils/third-party/azure"
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

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := azuredns.NewDefaultConfig()
	providerConfig.AuthMethod = "env"
	providerConfig.TenantID = config.TenantId
	providerConfig.ClientID = config.ClientId
	providerConfig.ClientSecret = config.ClientSecret
	providerConfig.SubscriptionID = config.SubscriptionId
	providerConfig.ResourceGroup = config.ResourceGroupName
	if config.CloudName != "" {
		env, err := xazure.GetCloudEnvConfiguration(config.CloudName)
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
