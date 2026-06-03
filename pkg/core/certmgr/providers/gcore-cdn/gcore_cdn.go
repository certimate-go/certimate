package gcorecdn

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gcore "github.com/G-Core/gcorelabscdn-go/gcore/provider"
	"github.com/G-Core/gcorelabscdn-go/sslcerts"

	"github.com/certimate-go/certimate/pkg/core"
	gcoresdk "github.com/certimate-go/certimate/pkg/sdk3rd/gcore"
)

type (
	Provider      = core.Certmgr
	UploadResult  = core.CertmgrUploadResult
	ReplaceResult = core.CertmgrReplaceResult
)

type CertmgrConfig struct {
	// G-Core API Token。
	ApiToken string `json:"apiToken"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *sslcerts.Service
}

var _ Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Certmgr{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (c *Certmgr) SetLogger(logger *slog.Logger) {
	if logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	} else {
		c.logger = logger
	}
}

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*UploadResult, error) {
	// 新增证书
	// REF: https://api.gcore.com/docs/cdn#tag/SSL-certificates/operation/add_ssl_certificates
	createCertificateReq := &sslcerts.CreateRequest{
		Name:           fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
		Cert:           certPEM,
		PrivateKey:     privkeyPEM,
		Automated:      false,
		ValidateRootCA: false,
	}
	createCertificateResp, err := c.sdkClient.Create(ctx, createCertificateReq)
	c.logger.Debug("sdk request 'sslcerts.Create'", slog.Any("request", createCertificateReq), slog.Any("response", createCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'sslcerts.Create': %w", err)
	}

	return &UploadResult{
		CertId:   fmt.Sprintf("%d", createCertificateResp.ID),
		CertName: createCertificateResp.Name,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*ReplaceResult, error) {
	return nil, core.ErrUnsupported
}

func createSDKClient(apiToken string) (*sslcerts.Service, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("gcore: invalid api token")
	}

	requester := gcore.NewClient(
		gcoresdk.BASE_URL,
		gcore.WithSigner(gcoresdk.NewAuthRequestSigner(apiToken)),
	)
	service := sslcerts.NewService(requester)
	return service, nil
}
