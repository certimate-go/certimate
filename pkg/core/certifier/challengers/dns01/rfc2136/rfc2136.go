package rfc2136

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/dnsupdate"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	Host                  string `json:"host"`
	Port                  int32  `json:"port"`
	TsigAlgorithm         string `json:"tsigAlgorithm,omitempty"`
	TsigKey               string `json:"tsigKey,omitempty"`
	TsigSecret            string `json:"tsigSecret,omitempty"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	if config.Port == 0 {
		config.Port = 53
	}

	if config.TsigAlgorithm == "" {
		config.TsigAlgorithm = "hmac-sha1."
	}

	providerConfig := dnsupdate.NewDefaultConfig()
	providerConfig.Nameserver = net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))
	providerConfig.TSIGAlgorithm = config.TsigAlgorithm
	providerConfig.TSIGKey = config.TsigKey
	providerConfig.TSIGSecret = config.TsigSecret
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := dnsupdate.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
