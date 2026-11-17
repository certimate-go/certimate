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
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xalibabacloud "github.com/certimate-go/certimate/pkg/utils/third-party/alibabacloud"
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
	// 是否自动移除同域名的其他证书。
	AutoPrune bool `json:"autoPrune,omitempty"`
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
		Region:          lo.Ternary(xalibabacloud.IsIntlRegion(config.Region), "ap-southeast-1", ""),
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

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询站点证书列表，并找出需要删除的同域名证书
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-listcertificates
	certIsAlreadySet := false
	certIdsToDelete := make([]string, 0)
	listCertificatesPageNumber := 1
	listCertificatesPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCertificatesReq := &aliesa.ListCertificatesRequest{
			SiteId: tea.Int64(d.config.SiteId),

			PageNumber: tea.Int64(int64(listCertificatesPageNumber)),
			PageSize:   tea.Int64(int64(listCertificatesPageSize)),
		}
		listCertificatesResp, err := d.sdkClient.ListCertificatesWithContext(ctx, listCertificatesReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'esa.ListCertificates'", slog.Any("request", listCertificatesReq), slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'esa.ListCertificates': %w", err)
		}

		for _, certItem := range listCertificatesResp.Body.Result {
			if tea.StringValue(certItem.CasId) == upres.CertId {
				certIsAlreadySet = true
				continue
			}

			certSANMatched := lo.ElementsMatch(certX509.DNSNames, strings.Split(tea.StringValue(certItem.SAN), ","))
			if certSANMatched {
				certIdsToDelete = append(certIdsToDelete, tea.StringValue(certItem.Id))
			}
		}

		if len(listCertificatesResp.Body.Result) < listCertificatesPageSize {
			break
		}

		listCertificatesPageNumber++
	}

	// 配置站点证书
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-setcertificate
	if certIsAlreadySet {
		d.logger.Info("no need to deploy esa site certificate")
		return &DeployResult{}, nil
	} else {
		certIdAsInt, _ := strconv.ParseInt(upres.CertId, 10, 64)
		setCertificateReq := &aliesa.SetCertificateRequest{
			Region: tea.String(d.config.Region),
			SiteId: tea.Int64(d.config.SiteId),
			Type:   tea.String("cas"),
			CasId:  tea.Int64(certIdAsInt),
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
	}

	// 删除站点证书
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-deletecertificate
	if d.config.AutoPrune && len(certIdsToDelete) > 0 {
		d.logger.Info("found esa site certificates to delete", slog.Any("certIds", certIdsToDelete))

		if err := xloop.ForRangeAllWithContext(ctx, certIdsToDelete, func(ctx context.Context, certId string, _ int) error {
			deleteCertificateReq := &aliesa.DeleteCertificateRequest{
				SiteId: tea.Int64(d.config.SiteId),
				Id:     tea.String(certId),
			}
			deleteCertificateResp, err := d.sdkClient.DeleteCertificateWithContext(ctx, deleteCertificateReq, &dara.RuntimeOptions{})
			d.logger.Debug("sdk request 'esa.DeleteCertificate'", slog.Any("request", deleteCertificateReq), slog.Any("response", deleteCertificateResp))
			if err != nil {
				if sdkErr, ok := err.(*tea.SDKError); ok {
					if sdkErrCode := tea.StringValue(sdkErr.Code); sdkErrCode == "Certificate.NotFound" {
						return nil
					}
				}

				return fmt.Errorf("failed to execute sdk request 'esa.DeleteCertificate': %w", err)
			}

			return nil
		}); err != nil {
			return nil, err
		}
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
