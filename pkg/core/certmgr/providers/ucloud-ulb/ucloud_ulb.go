package ucloudulb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ulb"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 优刻得地域。
	Region string `json:"region"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ucloudsdk.ULBClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId, config.Region)
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
	// 避免重复上传
	if upres, upok, err := c.tryGetResultIfCertExists(ctx, certPEM, privkeyPEM); err != nil {
		return nil, err
	} else if upok {
		c.logger.Info("ssl certificate already exists")
		return upres, nil
	}

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 生成新证书名（需符合优刻得命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())

	// 创建 SSL 证书
	// REF: https://docs.ucloud.cn/api/ulb-api/create_ssl
	createSSLReq := c.sdkClient.NewCreateSSLRequest()
	createSSLReq.SSLName = ucloud.String(certName)
	createSSLReq.SSLType = ucloud.String("Pem")
	createSSLReq.UserCert = ucloud.String(serverCertPEM)
	createSSLReq.CaCert = ucloud.String(intermediaCertPEM)
	createSSLReq.PrivateKey = ucloud.String(privkeyPEM)
	createSSLResp, err := c.sdkClient.CreateSSL(createSSLReq)
	c.logger.Debug("sdk request 'ulb.CreateSSL'", slog.Any("request", createSSLReq), slog.Any("response", createSSLResp))

	return &certmgr.UploadResult{
		CertId:   createSSLResp.SSLId,
		CertName: certName,
	}, nil
}

func (c *Certmgr) Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (*certmgr.OperateResult, error) {
	return nil, certmgr.ErrUnsupported
}

func (c *Certmgr) tryGetResultIfCertExists(ctx context.Context, certPEM, privkeyPEM string) (*certmgr.UploadResult, bool, error) {
	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, false, err
	}

	// 获取 SSL 证书信息
	// REF: https://docs.ucloud.cn/api/ulb-api/describe_ssl
	describeSSLOffset := 0
	describeSSLLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		describeSSLReq := c.sdkClient.NewDescribeSSLRequest()
		describeSSLReq.Offset = ucloud.Int(describeSSLOffset)
		describeSSLReq.Limit = ucloud.Int(describeSSLLimit)
		describeSSLResp, err := c.sdkClient.DescribeSSL(describeSSLReq)
		c.logger.Debug("sdk request 'ulb.DescribeSSL'", slog.Any("request", describeSSLReq), slog.Any("response", describeSSLResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'ulb.DescribeSSL': %w", err)
		}

		for _, sslItem := range describeSSLResp.DataSet {
			// 对比证书有效期
			if int64(sslItem.NotBefore) != certX509.NotBefore.Unix() || int64(sslItem.NotAfter) != certX509.NotAfter.Unix() {
				continue
			}

			// 对比证书及私钥内容
			// 按照“网站证书、私钥、中间证书”的方式拼接
			serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
			if err != nil {
				continue
			} else {
				oldSSLContent := sslItem.SSLContent
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\r", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\n", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\t", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, " ", "")

				newSSLContent := serverCertPEM + privkeyPEM + intermediaCertPEM
				newSSLContent = strings.ReplaceAll(newSSLContent, "\r", "")
				newSSLContent = strings.ReplaceAll(newSSLContent, "\n", "")
				newSSLContent = strings.ReplaceAll(newSSLContent, "\t", "")
				newSSLContent = strings.ReplaceAll(newSSLContent, " ", "")

				if oldSSLContent != newSSLContent {
					continue
				}
			}

			// 如果以上信息都一致，则视为已存在相同证书，直接返回
			return &certmgr.UploadResult{
				CertId:   sslItem.SSLId,
				CertName: sslItem.SSLName,
			}, true, nil
		}

		if len(describeSSLResp.DataSet) < describeSSLLimit {
			break
		}

		describeSSLOffset += describeSSLLimit
	}

	return nil, false, nil
}

func createSDKClient(privateKey, publicKey, projectId, region string) (*ucloudsdk.ULBClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId
	cfg.Region = region

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
