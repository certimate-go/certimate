package googleclouddns

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/gcloud"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"

	"github.com/certimate-go/certimate/pkg/core"
)

type ChallengerConfig struct {
	ServiceAccountKey     string `json:"serviceAccountKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	saKey := []byte(config.ServiceAccountKey)

	var saKeyJSON struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(saKey, &saKeyJSON); err != nil || saKeyJSON.ProjectID == "" {
		return nil, fmt.Errorf("googlecloud: project ID not found in service account key")
	}

	saConf, err := google.JWTConfigFromJSON(saKey, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to acquire service account config: %w", err)
	}

	providerConfig := gcloud.NewDefaultConfig()
	providerConfig.Project = saKeyJSON.ProjectID
	providerConfig.HTTPClient = saConf.Client(context.Background())
	if config.DnsPropagationTimeout != 0 {
		providerConfig.PropagationTimeout = time.Duration(config.DnsPropagationTimeout) * time.Second
	}
	if config.DnsTTL != 0 {
		providerConfig.TTL = config.DnsTTL
	}

	provider, err := gcloud.NewDNSProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
