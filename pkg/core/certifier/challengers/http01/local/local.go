package local

import (
	"fmt"

	"github.com/go-acme/lego/v4/providers/http/webroot"

	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	// 网站根目录路径。
	WebRootPath string `json:"webRootPath"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	provider, err := webroot.NewHTTPProvider(config.WebRootPath)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
