package yandexcloud

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/yandexcloud"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	FolderId              string `json:"folderId"`
	ServiceAccountKey     string `json:"serviceAccountKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := yandexcloud.NewDefaultConfig()
	providerConfig.FolderID = config.FolderId
	providerConfig.IamToken = base64.StdEncoding.EncodeToString([]byte(config.ServiceAccountKey))
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := yandexcloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
