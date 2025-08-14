package acmedns

import (
	"errors"
	"net/url"

	"github.com/go-acme/lego/v4/providers/dns/acmedns"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengeProviderConfig struct {
	ApiBase        string `json:"apiBase,omitempty"`
	StorageBaseUrl string `json:"storageBaseUrl,omitempty"`
	StoragePath    string `json:"storagePath,omitempty"`
}

func NewChallengeProvider(config *ChallengeProviderConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	ApiBase, _ := url.Parse(config.ApiBase)
	providerConfig := acmedns.NewDefaultConfig()
	providerConfig.APIBase = ApiBase.String()
	providerConfig.StorageBaseURL = config.StorageBaseUrl
	providerConfig.StoragePath = config.StoragePath

	provider, err := acmedns.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
