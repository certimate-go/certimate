package regru

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/regru"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	Username              string `json:"username"`
	Password              string `json:"password"`
	MtlsCertificate       string `json:"mtlsCertificate,omitempty"`
	MtlsPrivateKey        string `json:"mtlsPrivateKey,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := regru.NewDefaultConfig()
	providerConfig.Username = config.Username
	providerConfig.Password = config.Password
	providerConfig.TLSCert = config.MtlsCertificate
	providerConfig.TLSKey = config.MtlsPrivateKey
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := regru.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
