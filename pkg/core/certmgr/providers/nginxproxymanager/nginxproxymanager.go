package nginxproxymanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	npmsdk "github.com/certimate-go/certimate/pkg/sdk3rd/nginxproxymanager"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// NPM 服务地址。
	ServerUrl string `json:"serverUrl"`
	// NPM API 认证方式。
	// 可取值 "password"、"token"。
	// 零值时默认值 [AUTH_METHOD_PASSWORD]。
	AuthMethod string `json:"authMethod,omitempty"`
	// NPM 用户名。
	Username string `json:"username,omitempty"`
	// NPM 密码。
	Password string `json:"password,omitempty"`
	// NPM API Token。
	ApiToken string `json:"apiToken,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *npmsdk.Client
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.AuthMethod, config.Username, config.Password, config.ApiToken, config.AllowInsecureConnections)
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
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 获取全部证书，避免重复上传
	listCertificatesReq := &npmsdk.NginxListCertificatesRequest{}
	listCertificatesResp, err := c.sdkClient.NginxListCertificates(listCertificatesReq)
	c.logger.Debug("sdk request 'nginx.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'nginx.ListCertificates': %w", err)
	} else {
		for _, certItem := range *listCertificatesResp {
			if certItem.Meta.Certificate == serverCertPEM &&
				certItem.Meta.CertificateKey == privkeyPEM &&
				certItem.Meta.IntermediateCertificate == intermediaCertPEM {
				// 如果已存在相同证书，直接返回
				c.logger.Info("ssl certificate already exists")
				return &certmgr.UploadResult{
					CertId:   fmt.Sprintf("%d", certItem.Id),
					CertName: certItem.NiceName,
				}, nil
			}
		}
	}

	// 创建证书
	nginxCreateCertificateReq := &npmsdk.NginxCreateCertificateRequest{
		NiceName: fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
		Provider: "other",
	}
	nginxCreateCertificateResp, err := c.sdkClient.NginxCreateCertificate(nginxCreateCertificateReq)
	c.logger.Debug("sdk request 'nginx.CreateCertificate'", slog.Any("request", nginxCreateCertificateReq), slog.Any("response", nginxCreateCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'nginx.CreateCertificate': %w", err)
	}

	// 上传证书文件
	ngincxUploadCertificateReq := &npmsdk.NginxUploadCertificateRequest{
		CertificateMeta: npmsdk.CertificateMeta{
			Certificate:             serverCertPEM,
			CertificateKey:          privkeyPEM,
			IntermediateCertificate: intermediaCertPEM,
		},
	}
	ngincxUploadCertificateResp, err := c.sdkClient.NginxUploadCertificate(nginxCreateCertificateResp.Id, ngincxUploadCertificateReq)
	c.logger.Debug("sdk request 'nginx.UploadCertificate'", slog.Int64("request.certId", nginxCreateCertificateResp.Id), slog.Any("request", ngincxUploadCertificateReq), slog.Any("response", ngincxUploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'nginx.UploadCertificate': %w", err)
	}

	return &certmgr.UploadResult{
		CertId:   fmt.Sprintf("%d", nginxCreateCertificateResp.Id),
		CertName: nginxCreateCertificateResp.NiceName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	certId, err := strconv.ParseInt(certIdOrName, 10, 64)
	if err != nil {
		return nil, err
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 上传证书文件
	ngincxUploadCertificateReq := &npmsdk.NginxUploadCertificateRequest{
		CertificateMeta: npmsdk.CertificateMeta{
			Certificate:             serverCertPEM,
			CertificateKey:          privkeyPEM,
			IntermediateCertificate: intermediaCertPEM,
		},
	}
	ngincxUploadCertificateResp, err := c.sdkClient.NginxUploadCertificate(certId, ngincxUploadCertificateReq)
	c.logger.Debug("sdk request 'nginx.UploadCertificate'", slog.Int64("request.certId", certId), slog.Any("request", ngincxUploadCertificateReq), slog.Any("response", ngincxUploadCertificateResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'nginx.UploadCertificate': %w", err)
	}

	return &certmgr.OperateResult{}, nil
}

func createSDKClient(serverUrl, authMethod, username, password, apiToken string, skipTlsVerify bool) (*npmsdk.Client, error) {
	var client *npmsdk.Client
	var err error

	switch authMethod {
	case "", AUTH_METHOD_PASSWORD:
		{
			client, err = npmsdk.NewClient(serverUrl, username, password)
		}

	case AUTH_METHOD_TOKEN:
		{
			client, err = npmsdk.NewClientWithJwtToken(serverUrl, apiToken)
		}
	}

	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
