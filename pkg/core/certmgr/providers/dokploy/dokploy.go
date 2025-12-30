package dokploy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"
	
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	dokploysdk "github.com/certimate-go/certimate/pkg/sdk3rd/dokploy"
)

type CertmgrConfig struct {
	// Dokploy 服务地址。
	ServerUrl string `json:"serverUrl"`
	// Dokploy API Key。
	ApiKey string `json:"apiKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *dokploysdk.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiKey, config.AllowInsecureConnections)
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

func (c *Certmgr) Upload(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, error) {
	// 查询证书列表，避免重复上传
	// REF: https://docs.dokploy.com/docs/api/certificates#certificates-all
	certificatesAllReq := &dokploysdk.CertificatesAllRequest{}
	certificatesAllResp, err := c.sdkClient.CertificatesAll(certificatesAllReq)
	c.logger.Debug("sdk request 'certificates.all'", slog.Any("request", certificatesAllReq), slog.Any("response", certificatesAllResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificates.all': %w", err)
	} else {
		for _, certItem := range *certificatesAllResp {
			if certItem.CertificateData == certPEM && certItem.PrivateKey == privkeyPEM {
				// 如果已存在相同证书，直接返回
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   certItem.CertificateId,
					CertName: certItem.Name,
				}, nil
			}
		}
	}

	// 获取账号信息，找到默认的组织 ID
	// REF: https://docs.dokploy.com/docs/api/reference-user#user.get
	userGetReq := &dokploysdk.UserGetRequest{}
	userGetResp, err := c.sdkClient.UserGet(userGetReq)
	c.logger.Debug("sdk request 'user.get'", slog.Any("request", userGetReq), slog.Any("response", userGetResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'user.get': %w", err)
	}

	// 创建证书
	// REF: https://docs.dokploy.com/docs/api/certificates#certificates-create
	certificatesCreateReq := &dokploysdk.CertificatesCreateRequest{
		Name:            lo.ToPtr(fmt.Sprintf("certimate-%d", time.Now().Unix())),
		CertificateData: lo.ToPtr(certPEM),
		PrivateKey:      lo.ToPtr(privkeyPEM),
		OrganizationId:  lo.ToPtr(userGetResp.OrganizationId),
	}
	certificatesCreateResp, err := c.sdkClient.CertificatesCreate(certificatesCreateReq)
	c.logger.Debug("sdk request 'certificates.create'", slog.Any("request", certificatesCreateReq), slog.Any("response", certificatesCreateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'certificates.create': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   certificatesCreateResp.CertificateId,
		CertName: certificatesCreateResp.Name,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func createSDKClient(serverUrl, apiKey string, skipTlsVerify bool) (*dokploysdk.Client, error) {
	client, err := dokploysdk.NewClient(serverUrl, apiKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
