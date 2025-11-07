package edgedns

import (
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/edgegrid"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/edgedns"
)

// ChallengeProviderConfig is the configuration for the EdgeDNS provider
type ChallengeProviderConfig struct {
	ClientToken           string
	ClientSecret          string
	AccessToken           string
	Host                  string
	DnsPropagationTimeout time.Duration
	DnsTTL                int
}

func NewChallengeProvider(config *ChallengeProviderConfig) (challenge.Provider, error) {
	edgegridConfig := &edgegrid.Config{
		Host:         config.Host,
		ClientToken:  config.ClientToken,
		ClientSecret: config.ClientSecret,
		AccessToken:  config.AccessToken,
		MaxBody:      131072,
		HeaderToSign: []string{
			"X-Akamai-ACS-Action",
			"X-Akamai-ACS-Auth-Data",
			"X-Akamai-ACS-Auth-Sign",
		},
	}

	providerConfig := edgedns.NewDefaultConfig()
	providerConfig.Config = edgegridConfig

	if config.DnsPropagationTimeout > 0 {
		providerConfig.PropagationTimeout = config.DnsPropagationTimeout
	}
	if config.DnsTTL > 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := edgedns.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
