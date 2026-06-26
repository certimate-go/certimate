package googlecloud

import (
	"context"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/providers/dns/gcloud"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"

	"github.com/certimate-go/certimate/pkg/core"
	xgcp "github.com/certimate-go/certimate/pkg/utils/third-party/gcp"
)

type ChallengerConfig struct {
	ProjectId             string `json:"projectId"`
	ServiceAccountKey     string `json:"serviceAccountKey"`
	DnsPropagationTimeout int    `json:"dnsPropagationTimeout,omitempty"`
	DnsTTL                int    `json:"dnsTTL,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	projectId, err := xgcp.GetProjectIDFromServiceAccountKey(config.ServiceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: %w", err)
	} else if projectId != config.ProjectId {
		return nil, fmt.Errorf("googlecloud: invalid project ID: expected '%s', got '%s'", config.ProjectId, projectId)
	}

	saKey := []byte(config.ServiceAccountKey)
	saConf, err := google.JWTConfigFromJSON(saKey, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to acquire service account config: %w", err)
	}

	providerConfig := gcloud.NewDefaultConfig()
	providerConfig.Project = projectId
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
