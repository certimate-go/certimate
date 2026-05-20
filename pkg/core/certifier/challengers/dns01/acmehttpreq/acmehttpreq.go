package acmehttpreq

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/httpreq"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	Endpoint              string `json:"endpoint"`
	Mode                  string `json:"mode"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	endpoint, _ := url.Parse(config.Endpoint)
	providerConfig := httpreq.NewDefaultConfig()
	providerConfig.Endpoint = endpoint
	providerConfig.Mode = config.Mode
	providerConfig.Username = config.Username
	providerConfig.Password = config.Password
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}

	provider, err := httpreq.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
