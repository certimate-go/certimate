package oraclecloud

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/oraclecloud"
	"github.com/nrdcg/oci-go-sdk/common/v1065"
	"github.com/nrdcg/oci-go-sdk/common/v1065/auth"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	AuthMethod            string `json:"authMethod"`
	PrivateKey            string `json:"privateKey,omitempty"`
	PrivateKeyPassphrase  string `json:"privateKeyPassphrase,omitempty"`
	PublicKeyFingerprint  string `json:"publicKeyFingerprint,omitempty"`
	TenancyOcid           string `json:"tenancyOcid,omitempty"`
	UserOcid              string `json:"userOcid,omitempty"`
	Region                string `json:"region"`
	CompartmentOcid       string `json:"compartmentOcid"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := oraclecloud.NewDefaultConfig()
	providerConfig.CompartmentID = config.CompartmentOcid
	switch config.AuthMethod {
	case AUTH_METHOD_APIKEY:
		pkpwd := (*string)(nil)
		if config.PrivateKeyPassphrase != "" {
			pkpwd = &config.PrivateKeyPassphrase
		}
		providerConfig.OCIConfigProvider = common.NewRawConfigurationProvider(
			config.TenancyOcid,
			config.UserOcid,
			config.Region,
			config.PublicKeyFingerprint,
			config.PrivateKey,
			pkpwd,
		)
	case AUTH_METHOD_INSTANCEPRINCIPAL:
		configurationProvider, err := auth.InstancePrincipalConfigurationProviderForRegion(common.Region(config.Region))
		if err != nil {
			return nil, fmt.Errorf("oraclecloud: %w", err)
		}
		providerConfig.OCIConfigProvider = configurationProvider
	case AUTH_METHOD_RESOURCEPRINCIPAL:
		configurationProvider, err := auth.ResourcePrincipalConfigurationProviderForRegion(common.Region(config.Region))
		if err != nil {
			return nil, fmt.Errorf("oraclecloud: %w", err)
		}
		providerConfig.OCIConfigProvider = configurationProvider
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", config.AuthMethod)
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := oraclecloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
