package ucloudulb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/upathx"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type CertmgrConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
}

type Certmgr struct {
	config    *CertmgrConfig
	logger    *slog.Logger
	sdkClient *ucloudsdk.UPathXClient
}

var _ certmgr.Provider = (*Certmgr)(nil)

func NewCertmgr(config *CertmgrConfig) (*Certmgr, error) {
	if config == nil {
		return nil, errors.New("the configuration of the certmgr provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId)
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
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 创建证书
	// REF: https://docs.ucloud.cn/api/pathx-api/create_path_xssl
	createPathXSSLReq := c.sdkClient.NewCreatePathXSSLRequest()
	createPathXSSLReq.SSLName = ucloud.String(certName)
	createPathXSSLReq.SSLType = ucloud.String("Pem")
	createPathXSSLReq.UserCert = ucloud.String(serverCertPEM)
	createPathXSSLReq.CACert = ucloud.String(intermediaCertPEM)
	createPathXSSLReq.PrivateKey = ucloud.String(privkeyPEM)
	createPathXSSLResp, err := c.sdkClient.CreatePathXSSL(createPathXSSLReq)
	c.logger.Debug("sdk request 'pathx.CreatePathXSSL'", slog.Any("request", createPathXSSLReq), slog.Any("response", createPathXSSLResp))

	return &certmgr.UploadResult{
		CertId:   createPathXSSLResp.SSLId,
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

	// 获取证书信息
	// REF: https://docs.ucloud.cn/api/pathx-api/describe_path_xssl
	describePathXSSLOffset := 0
	describePathXSSLLimit := 100
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}

		describePathXSSLReq := c.sdkClient.NewDescribePathXSSLRequest()
		describePathXSSLReq.Offset = ucloud.Int(describePathXSSLOffset)
		describePathXSSLReq.Limit = ucloud.Int(describePathXSSLLimit)
		describePathXSSLResp, err := c.sdkClient.DescribePathXSSL(describePathXSSLReq)
		c.logger.Debug("sdk request 'pathx.DescribePathXSSL'", slog.Any("request", describePathXSSLReq), slog.Any("response", describePathXSSLResp))
		if err != nil {
			return nil, false, fmt.Errorf("failed to execute sdk request 'pathx.DescribePathXSSL': %w", err)
		}

		for _, sslItem := range describePathXSSLResp.DataSet {
			// 对比证书有效期
			if int64(sslItem.ExpireTime) != certX509.NotAfter.Unix() {
				continue
			}

			// 对比证书及私钥内容
			// 按照“私钥、证书链”的方式拼接
			serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
			if err != nil {
				continue
			} else {
				oldSSLContent := sslItem.SSLContent
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\r", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\n", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, "\t", "")
				oldSSLContent = strings.ReplaceAll(oldSSLContent, " ", "")

				newSSLContent := privkeyPEM + serverCertPEM + intermediaCertPEM
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

		if len(describePathXSSLResp.DataSet) < describePathXSSLLimit {
			break
		}

		describePathXSSLOffset += describePathXSSLLimit
	}

	return nil, false, nil
}

func createSDKClient(privateKey, publicKey, projectId string) (*ucloudsdk.UPathXClient, error) {
	if privateKey == "" {
		return nil, errors.New("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, errors.New("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId

	// PathX 相关接口要求必传 ProjectId 参数
	if cfg.ProjectId == "" {
		defaultProjectId, err := getSDKDefaultProjectId(privateKey, publicKey)
		if err != nil {
			return nil, err
		}

		cfg.ProjectId = defaultProjectId
	}

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}

func getSDKDefaultProjectId(privateKey, publicKey string) (string, error) {
	cfg := ucloud.NewConfig()

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := uaccount.NewClient(&cfg, &credential)

	request := client.NewGetProjectListRequest()
	response, err := client.GetProjectList(request)
	if err != nil {
		return "", err
	}

	for _, projectItem := range response.ProjectSet {
		if projectItem.IsDefault {
			return projectItem.ProjectId, nil
		}
	}

	return "", errors.New("ucloud: no default project found")
}
