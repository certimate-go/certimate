package certmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
)

type PurgeCertificateRequest struct {
	// 提供商相关
	_                      struct{}
	Provider               domain.PurgeProviderType
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 清理相关
	_      struct{}
	Expiry time.Duration
}

type PurgeCertificateResponse struct{}

func (c *Client) PurgeCertificate(ctx context.Context, request *PurgeCertificateRequest) (*PurgeCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	providerFactory, err := purgers.Registries.Get(request.Provider)
	if err != nil {
		return nil, err
	}

	provider, err := providerFactory(&purgers.ProviderFactoryOptions{
		ProviderAccessConfig:   request.ProviderAccessConfig,
		ProviderExtendedConfig: request.ProviderExtendedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize purge provider '%s': %w", request.Provider, err)
	}

	provider.SetLogger(c.logger)
	if _, err := provider.Purge(ctx, request.Expiry); err != nil {
		return nil, err
	}

	return &PurgeCertificateResponse{}, nil
}
