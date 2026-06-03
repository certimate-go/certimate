package jdcloudwaf

import (
	"context"
	"fmt"
	"log/slog"

	jdcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	jdwafapis "github.com/jdcloud-api/jdcloud-sdk-go/services/waf/apis"
	jdwafmodels "github.com/jdcloud-api/jdcloud-sdk-go/services/waf/models"

	jdwaf "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/jdcloud-api/jdcloud-sdk-go/services/waf/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/jdcloud-ssl"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 京东云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 京东云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 京东云地域 ID。
	RegionId string `json:"regionId"`
	// WAF 实例 ID。
	InstanceId string `json:"instanceId"`
	// 防护域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *jdwaf.WafClient
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 绑定证书
	// REF: https://docs.jdcloud.com/cn/web-application-firewall/api/bindcert
	bindCertReq := jdwafapis.NewBindCertRequestWithoutParam()
	bindCertReq.SetRegionId(d.config.RegionId)
	bindCertReq.SetWafInstanceId(d.config.InstanceId)
	bindCertReq.SetReq(&jdwafmodels.AssignCertReq{
		WafInstanceId: d.config.InstanceId,
		Domain:        d.config.Domain,
		CertId:        upres.CertId,
	})
	bindCertResp, err := d.sdkClient.BindCert(bindCertReq)
	d.logger.Debug("sdk request 'waf.BindCert'", slog.Any("request", bindCertReq), slog.Any("response", bindCertResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'waf.BindCert': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret string) (*jdwaf.WafClient, error) {
	clientCredentials := jdcore.NewCredentials(accessKeyId, accessKeySecret)
	client := jdwaf.NewWafClient(clientCredentials)
	client.DisableLogger()
	return client, nil
}
