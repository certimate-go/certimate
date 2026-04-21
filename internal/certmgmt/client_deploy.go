package certmgmt

import (
	"context"
	"errors"
	"fmt"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
)

type DeployCertificateRequest struct {
	// 提供商相关
	Provider               string
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 证书相关
	Certificate       string
	PrivateKey        string
	IssuerCertificate string
}

type DeployCertificateResponse struct{}

func (c *Client) DeployCertificate(ctx context.Context, request *DeployCertificateRequest) (*DeployCertificateResponse, error) {
	if request == nil {
		return nil, errors.New("the request is nil")
	}

	providerFactory, err := deployers.Registries.Get(domain.DeploymentProviderType(request.Provider))
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
	if _, err := provider.Deploy(ctx, request.Certificate, request.PrivateKey, request.IssuerCertificate); err != nil {
		return nil, err
	}

	return &DeployCertificateResponse{}, nil
}
