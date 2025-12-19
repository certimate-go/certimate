package uclouduewaf

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/uewaf"
)

type DeployerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 自定义域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ucloudsdk.UEWAFClient
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.Domain == "" {
		return nil, errors.New("config `domain` is required")
	}

	// 生成优刻得所需的证书参数
	certPEMBase64 := base64.StdEncoding.EncodeToString([]byte(certPEM))
	privkeyPEMBase64 := base64.StdEncoding.EncodeToString([]byte(privkeyPEM))
	certMd5 := md5.Sum([]byte(certPEMBase64 + privkeyPEMBase64))
	certMd5Hex := hex.EncodeToString(certMd5[:])
	certName := fmt.Sprintf("certimate_%d", time.Now().UnixMilli())

	// 添加 SSL 证书
	// REF: https://docs.ucloud.cn/api/uewaf-api/add_waf_domain_certificate_info
	addWafDomainCertificateInfoReq := d.sdkClient.NewAddWafDomainCertificateInfoRequest()
	addWafDomainCertificateInfoReq.Domain = ucloud.String(d.config.Domain)
	addWafDomainCertificateInfoReq.CertificateName = ucloud.String(certName)
	addWafDomainCertificateInfoReq.SslPublicKey = ucloud.String(certPEMBase64)
	addWafDomainCertificateInfoReq.SslPrivateKey = ucloud.String(privkeyPEMBase64)
	addWafDomainCertificateInfoReq.SslMD = ucloud.String(certMd5Hex)
	addWafDomainCertificateInfoResp, err := d.sdkClient.AddWafDomainCertificateInfo(addWafDomainCertificateInfoReq)
	d.logger.Debug("sdk request 'uewaf.AddWafDomainCertificateInfo'", slog.Any("request", addWafDomainCertificateInfoReq), slog.Any("response", addWafDomainCertificateInfoResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'uewaf.AddWafDomainCertificateInfo': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(privateKey, publicKey, projectId string) (*ucloudsdk.UEWAFClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
