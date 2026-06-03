package flexcdn

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	flexcdnsdk "github.com/certimate-go/certimate/pkg/sdk3rd/flexcdn"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// FlexCDN 服务地址。
	ServerUrl string `json:"serverUrl"`
	// FlexCDN 用户角色。
	// 可取值 "user"、"admin"。
	ApiRole string `json:"apiRole"`
	// FlexCDN AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// FlexCDN AccessKey。
	AccessKey string `json:"accessKey"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *flexcdnsdk.Client
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiRole, config.AccessKeyId, config.AccessKey, config.AllowInsecureConnections)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == 0 {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	// 修改证书
	// REF: https://flexcdn.cloud/dev/api/service/SSLCertService?role=user#updateSSLCert
	updateSSLCertReq := &flexcdnsdk.UpdateSSLCertRequest{
		SSLCertId:   d.config.CertificateId,
		IsOn:        true,
		Name:        fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
		Description: "upload from certimate",
		ServerName:  certX509.Subject.CommonName,
		IsCA:        false,
		CertData:    base64.StdEncoding.EncodeToString([]byte(certPEM)),
		KeyData:     base64.StdEncoding.EncodeToString([]byte(privkeyPEM)),
		TimeBeginAt: certX509.NotBefore.Unix(),
		TimeEndAt:   certX509.NotAfter.Unix(),
		DNSNames:    certX509.DNSNames,
		CommonNames: []string{certX509.Subject.CommonName},
	}
	updateSSLCertResp, err := d.sdkClient.UpdateSSLCertWithContext(ctx, updateSSLCertReq)
	d.logger.Debug("sdk request 'UpdateSSLCert'", slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'UpdateSSLCert': %w", err)
	}

	return nil
}

func createSDKClient(serverUrl, apiRole, accessKeyId, accessKey string, skipTlsVerify bool) (*flexcdnsdk.Client, error) {
	client, err := flexcdnsdk.NewClient(serverUrl, apiRole, accessKeyId, accessKey)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
