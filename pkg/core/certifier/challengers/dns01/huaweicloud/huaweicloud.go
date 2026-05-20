package huaweicloud

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	hwc "github.com/go-acme/lego/v4/providers/dns/huaweicloud"
)

type ChallengerConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	SecretAccessKey       string `json:"secretAccessKey"`
	Region                string `json:"region"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	region := config.Region
	if region == "" {
		// 华为云的 SDK 要求必须传一个区域，实际上 DNS 服务用不到，但不传会报错
		region = "cn-north-1"
	}

	providerConfig := hwc.NewDefaultConfig()
	providerConfig.AccessKeyID = config.AccessKeyId
	providerConfig.SecretAccessKey = config.SecretAccessKey
	providerConfig.Region = region
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = int32(config.DnsTTL)
	}

	provider, err := hwc.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
