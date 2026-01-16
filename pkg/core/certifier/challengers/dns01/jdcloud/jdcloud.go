package jdcloud

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/jdcloud"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	AccessKeySecret       string `json:"accessKeySecret"`
	RegionId              string `json:"regionId"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	regionId := config.RegionId
	if regionId == "" {
		// 京东云的 SDK 要求必须传一个区域，实际上 DNS 服务用不到，但不传会报错
		regionId = "cn-north-1"
	}

	providerConfig := jdcloud.NewDefaultConfig()
	providerConfig.AccessKeyID = config.AccessKeyId
	providerConfig.AccessKeySecret = config.AccessKeySecret
	providerConfig.RegionID = regionId
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := jdcloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
