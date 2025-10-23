package gcorecdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/G-Core/gcorelabscdn-go/gcore/provider"
	"github.com/G-Core/gcorelabscdn-go/sslcerts"

	"github.com/certimate-go/certimate/pkg/core"
	gcoresdk "github.com/certimate-go/certimate/pkg/sdk3rd/gcore"
)

type SSLManagerProviderConfig struct {
	// G-Core API Token。
	ApiToken string `json:"apiToken"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *slog.Logger
	sdkClient *sslcerts.Service
}

var _ core.SSLManager = (*SSLManagerProvider)(nil)

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (m *SSLManagerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		m.logger = slog.New(slog.DiscardHandler)
	} else {
		m.logger = logger
	}
}

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLManageUploadResult, error) {
	// 新增证书
	// REF: https://api.gcore.com/docs/cdn#tag/SSL-certificates/operation/add_ssl_certificates
	createCertificateReq := &sslcerts.CreateRequest{
		Name:           fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
		Cert:           certPEM,
		PrivateKey:     privkeyPEM,
		Automated:      false,
		ValidateRootCA: false,
	}
	createCertificateResp, err := m.sdkClient.Create(context.TODO(), createCertificateReq)
	m.logger.Debug("sdk request 'sslcerts.Create'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcerts.Create': %w", err)
	}

	return &core.SSLManageUploadResult{
		CertId:   fmt.Sprintf("%d", createCertificateResp.ID),
		CertName: createCertificateResp.Name,
	}, nil
}

func createSDKClient(apiToken string) (*sslcerts.Service, error) {
	if apiToken == "" {
		return nil, errors.New("invalid gcore api token")
	}

	requester := provider.NewClient(
		gcoresdk.BASE_URL,
		provider.WithSigner(gcoresdk.NewAuthRequestSigner(apiToken)),
	)
	service := sslcerts.NewService(requester)
	return service, nil
}
