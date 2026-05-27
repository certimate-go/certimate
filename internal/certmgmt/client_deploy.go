package certmgmt

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
)

type DeployCertificateRequest struct {
	// 提供商相关
	Provider               domain.DeploymentProviderType
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 证书相关
	CertificatePEM string
	PrivateKeyPEM  string
}

type DeployCertificateResponse struct{}

func (c *Client) DeployCertificate(ctx context.Context, request *DeployCertificateRequest) (*DeployCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	providerFactory, err := deployers.Registries.Get(request.Provider)
	if err != nil {
		return nil, err
	}

	provider, err := providerFactory(&deployers.ProviderFactoryOptions{
		ProviderAccessConfig:   request.ProviderAccessConfig,
		ProviderExtendedConfig: request.ProviderExtendedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deployment provider '%s': %w", request.Provider, err)
	}

	provider.SetLogger(c.logger)
	if _, err := provider.Deploy(ctx, request.CertificatePEM, request.PrivateKeyPEM); err != nil {
		return nil, err
	}

	return &DeployCertificateResponse{}, nil
}
