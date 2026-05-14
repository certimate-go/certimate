package awsroute53

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/route53"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"secretAccessKey"`
	Region                string `json:"region"`
	HostedZoneId          string `json:"hostedZoneId,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := route53.NewDefaultConfig()
	providerConfig.AccessKeyID = config.AccessKeyId
	providerConfig.SecretAccessKey = config.SecretAccessKey
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
