package local

import (
	"fmt"

	"github.com/go-acme/lego/v5/providers/http/webroot"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	// 网站根目录路径。
	WebRootPath string `json:"webRootPath"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	if config.WebRootPath == "" {
		return nil, fmt.Errorf("local: webroot path must be set")
	}

	provider, err := webroot.NewHTTPProvider(config.WebRootPath)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
