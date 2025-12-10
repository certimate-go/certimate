package cpanel

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/cpanel"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

type ChallengerConfig struct {
	ServerUrl                string `json:"serverUrl"`
	Username                 string `json:"username"`
	ApiToken                 string `json:"apiToken"`
	DnsPropagationTimeout    int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                   int    `json:"dnsTTL,omitempty"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	providerConfig := cpanel.NewDefaultConfig()
	providerConfig.Mode = "cpanel"
	providerConfig.BaseURL = config.ServerUrl
	providerConfig.Username = config.Username
	providerConfig.Token = config.ApiToken
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

	provider, err := cpanel.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
