package ovhcloud

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/ovh"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	Endpoint              string `json:"endpoint"`
	AuthMethod            string `json:"authMethod"`
	ApplicationKey        string `json:"applicationKey,omitempty"`
	ApplicationSecret     string `json:"applicationSecret,omitempty"`
	ConsumerKey           string `json:"consumerKey,omitempty"`
	ClientId              string `json:"clientId,omitempty"`
	ClientSecret          string `json:"clientSecret,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := ovh.NewDefaultConfig()
	providerConfig.APIEndpoint = config.Endpoint
	switch config.AuthMethod {
	case AUTH_METHOD_APPLICATION:
		providerConfig.ApplicationKey = config.ApplicationKey
		providerConfig.ApplicationSecret = config.ApplicationSecret
		providerConfig.ConsumerKey = config.ConsumerKey
	case AUTH_METHOD_OAUTH2:
		providerConfig.OAuth2Config = &ovh.OAuth2Config{
			ClientID:     config.ClientId,
			ClientSecret: config.ClientSecret,
		}
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", config.AuthMethod)
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := ovh.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
