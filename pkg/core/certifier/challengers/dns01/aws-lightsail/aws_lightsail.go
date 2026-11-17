package awslightsail

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"

	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/aws-lightsail/internal"
)

type ChallengerConfig struct {
	AuthMethod            string `json:"authMethod"`
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"secretAccessKey"`
	Region                string `json:"region"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	switch config.AuthMethod {
	case "":
		if config.AccessKeyId != "" && config.SecretAccessKey != "" {
			providerConfig.AWSCredentialsProvider = credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.SecretAccessKey, "")
		}
	case AUTH_METHOD_ACCESSKEY:
		providerConfig.AWSCredentialsProvider = credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.SecretAccessKey, "")
	case AUTH_METHOD_IMDS:
		providerConfig.AWSCredentialsProvider = aws.NewCredentialsCache(ec2rolecreds.New())
	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", config.AuthMethod)
	}
	providerConfig.Region = config.Region
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}

	provider, err := internal.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
