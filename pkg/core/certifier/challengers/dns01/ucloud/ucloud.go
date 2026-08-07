package ucloud

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/ucloud/internal"
)

type ChallengerConfig struct {
	PrivateKey            string `json:"privateKey"`
	PublicKey             string `json:"publicKey"`
	ProjectId             string `json:"projectId,omitempty"`
	Endpoint              string `json:"endpoint,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.PrivateKey = config.PrivateKey
	providerConfig.PublicKey = config.PublicKey
	providerConfig.ProjectID = config.ProjectId
	providerConfig.BaseURL = config.Endpoint
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}

	provider, err := internal.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
