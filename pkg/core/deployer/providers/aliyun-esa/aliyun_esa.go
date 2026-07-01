package aliyunesa

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	aliesa "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/esa-20240910/v3/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 阿里云 ESA 站点 ID。
	SiteId int64 `json:"siteId"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *aliesa.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region: lo.
			If(config.Region == "" || strings.HasPrefix(config.Region, "cn-"), "cn-hangzhou").
			Else("ap-southeast-1"),
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
	if d.config.SiteId == 0 {
		return nil, fmt.Errorf("config `siteId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 配置站点证书
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-setcertificate
	certId, _ := strconv.ParseInt(upres.CertId, 10, 64)
	setCertificateReq := &aliesa.SetCertificateRequest{
		Region: tea.String(d.config.Region),
		SiteId: tea.Int64(d.config.SiteId),
		Type:   tea.String("cas"),
		CasId:  tea.Int64(certId),
	}
	setCertificateResp, err := d.sdkClient.SetCertificateWithContext(ctx, setCertificateReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'esa.SetCertificate'", slog.Any("request", setCertificateReq), slog.Any("response", setCertificateResp))
	if err != nil {
		if sdkErr, ok := err.(*tea.SDKError); ok {
			if sdkErrCode := tea.StringValue(sdkErr.Code); sdkErrCode == "Certificate.Duplicated" {
				return &DeployResult{}, nil
			}
		}

		return nil, fmt.Errorf("failed to execute sdk request 'esa.SetCertificate': %w", err)
	}

	return &DeployResult{}, nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*aliesa.Client, error) {
	// 接入点一览 https://api.aliyun.com/product/ESA
	var endpoint string
	switch region {
	case "":
		endpoint = "esa.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("esa.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := aliesa.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
