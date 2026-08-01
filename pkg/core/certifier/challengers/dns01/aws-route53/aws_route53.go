package awsroute53

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/route53"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	AuthMethod            string `json:"authMethod"`
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"secretAccessKey"`
	Region                string `json:"region"`
	HostedZoneId          string `json:"hostedZoneId,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := route53.NewDefaultConfig()
	switch config.AuthMethod {
	case "":
		if config.AccessKeyId != "" && config.SecretAccessKey != "" {
			providerConfig.AccessKeyID = config.AccessKeyId
			providerConfig.SecretAccessKey = config.SecretAccessKey
		}
	case AUTH_METHOD_ACCESSKEY:
		providerConfig.AccessKeyID = config.AccessKeyId
		providerConfig.SecretAccessKey = config.SecretAccessKey
		providerConfig.AssumeRoleArn = ""
		providerConfig.ExternalID = ""
	case AUTH_METHOD_IMDS:
		providerConfig.AccessKeyID = ""
		providerConfig.SecretAccessKey = ""
		providerConfig.AssumeRoleArn = ""
		providerConfig.ExternalID = ""
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", config.AuthMethod)
	}
	providerConfig.Region = config.Region
	if config.HostedZoneId != "" {
		providerConfig.HostedZoneID = config.HostedZoneId
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := route53.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
