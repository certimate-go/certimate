package certmgmt

import (
	"context"
	"fmt"

	"github.com/certimate-go/certimate/internal/certmgmt/certsources"
	"github.com/certimate-go/certimate/internal/domain"
)

type FetchCertificateRequest struct {
	// 提供商相关
	Provider               domain.CertsourceProviderType
	ProviderAccessConfig   map[string]any
	ProviderExtendedConfig map[string]any

	// 证书相关
	Domain string
	CertId string
}

type FetchCertificateResponse struct {
	CertPEM    string
	PrivkeyPEM string

	ExtendedData map[string]any
}

func (c *Client) FetchCertificate(ctx context.Context, request *FetchCertificateRequest) (*FetchCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	providerFactory, err := certsources.Registries.Get(request.Provider)
	if err != nil {
		return nil, err
	}

	provider, err := providerFactory(&certsources.ProviderFactoryOptions{
		ProviderAccessConfig:   request.ProviderAccessConfig,
		ProviderExtendedConfig: request.ProviderExtendedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize certificate source provider '%s': %w", request.Provider, err)
	}

	provider.SetLogger(c.logger)
	res, err := provider.Fetch(ctx, request.Domain, request.CertId)
	if err != nil {
		return nil, err
	}

	return &FetchCertificateResponse{
		CertPEM:      res.CertPEM,
		PrivkeyPEM:   res.PrivkeyPEM,
		ExtendedData: res.ExtendedData,
	}, nil
}
