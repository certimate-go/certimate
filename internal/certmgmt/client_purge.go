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

type PurgeCertificateResponse struct {
	Count int
}

func (c *Client) PurgeCertificate(ctx context.Context, request *PurgeCertificateRequest) (*PurgeCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}
	if request.Expiry < 0 {
		return nil, fmt.Errorf("the expiry must not be negative, got %s", request.Expiry)
	}

	provider := domain.ResolvePurgeProviderAlias(request.Provider)
	providerFactory, err := purgers.Registries.Get(provider)
	if err != nil {
		return nil, err
	}

	providerImpl, err := providerFactory(&purgers.ProviderFactoryOptions{
		ProviderAccessConfig:   request.ProviderAccessConfig,
		ProviderExtendedConfig: request.ProviderExtendedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize purge provider '%s': %w", provider, err)
	}

	providerImpl.SetLogger(c.logger)
	purgeRes, err := providerImpl.Purge(ctx, request.Expiry)
	if err != nil {
		return nil, err
	}
	if purgeRes == nil {
		return nil, fmt.Errorf("purge provider '%s' returned no result", provider)
	}

	return &PurgeCertificateResponse{Count: purgeRes.Count}, nil
}
