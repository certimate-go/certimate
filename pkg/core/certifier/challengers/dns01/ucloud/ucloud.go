package ucloud

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/ucloud"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	PrivateKey            string `json:"privateKey"`
	PublicKey             string `json:"publicKey"`
	ProjectId             string `json:"projectId,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	providerConfig := ucloud.NewDefaultConfig()
	providerConfig.PrivateKey = config.PrivateKey
	providerConfig.PublicKey = config.PublicKey
	providerConfig.ProjectID = config.ProjectId
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}

	provider, err := ucloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
