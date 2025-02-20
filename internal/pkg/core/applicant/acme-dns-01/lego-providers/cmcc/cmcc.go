package cmcc

import (
	"errors"
	"time"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/usual2970/certimate/internal/pkg/core/applicant/acme-dns-01/lego-providers/cmcc/internal"
)

type CmccApplicantConfig struct {
	AccessKeyId           string `json:"accessKeyId"`
	AccessKeySecret       string `json:"accessKeySecret"`
	DnsPropagationTimeout int32  `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int32  `json:"dnsTTL,omitempty"`
}

func NewChallengeProvider(config *CmccApplicantConfig) (challenge.Provider, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.AccessKeyId = config.AccessKeyId
	providerConfig.AccessKeySecret = config.AccessKeySecret
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
