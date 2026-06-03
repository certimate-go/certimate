package cloudflare

import (
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/cloudflare"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	DnsApiToken           string `json:"dnsApiToken"`
	ZoneApiToken          string `json:"zoneApiToken,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := cloudflare.NewDefaultConfig()
	providerConfig.AuthToken = config.DnsApiToken
	providerConfig.ZoneToken = config.ZoneApiToken
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := cloudflare.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
