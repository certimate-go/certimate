package technitiumdns

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/technitium"

	"github.com/certimate-go/certimate/pkg/core"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

type ChallengerConfig struct {
	ServerUrl                string `json:"serverUrl"`
	ApiToken                 string `json:"apiToken"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
	DnsPropagationTimeout    int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                   int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := technitium.NewDefaultConfig()
	providerConfig.BaseURL = config.ServerUrl
	providerConfig.APIToken = config.ApiToken
	if config.AllowInsecureConnections {
		transport := xhttp.NewDefaultTransport()
		transport.DisableKeepAlives = true
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		providerConfig.HTTPClient.Transport = transport
	}
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := technitium.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
