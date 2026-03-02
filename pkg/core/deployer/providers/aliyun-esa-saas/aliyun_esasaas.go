package aliyunesasaas

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	aliesa "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-esa-saas/internal"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
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
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// SaaS 域名（不支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.EsaClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	if d.config.SiteId == 0 {
		return nil, errors.New("config `siteId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名 ID 列表
	var hostnameIds []int64
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			hostnameCandidates, err := d.getAllHostnames(ctx)
			if err != nil {
				return nil, err
			}

			hostname, ok := lo.Find(hostnameCandidates, func(hostname *aliesa.ListCustomHostnamesResponseBodyHostnames) bool {
				return d.config.Domain == tea.StringValue(hostname.Hostname)
			})
			if !ok {
				return nil, fmt.Errorf("could not find hostname '%s'", d.config.Domain)
			}

			hostnameIds = []int64{tea.Int64Value(hostname.HostnameId)}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			hostnameCandidates, err := d.getAllHostnames(ctx)
			if err != nil {
				return nil, err
			}

			hostnames := lo.Filter(hostnameCandidates, func(hostname *aliesa.ListCustomHostnamesResponseBodyHostnames, _ int) bool {
				if strings.HasPrefix(d.config.Domain, "*.") {
					return xcerthostname.IsMatch(d.config.Domain, tea.StringValue(hostname.Hostname))
				} else {
					return d.config.Domain == tea.StringValue(hostname.Hostname)
				}
			})
			if len(hostnames) == 0 {
				return nil, errors.New("could not find any hostnames matched by wildcard")
			}

			hostnameIds = lo.Map(hostnames, func(hostname *aliesa.ListCustomHostnamesResponseBodyHostnames, _ int) int64 {
				return tea.Int64Value(hostname.HostnameId)
			})
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			hostnameCandidates, err := d.getAllHostnames(ctx)
			if err != nil {
				return nil, err
			}

			hostnames := lo.Filter(hostnameCandidates, func(hostname *aliesa.ListCustomHostnamesResponseBodyHostnames, _ int) bool {
				return certX509.VerifyHostname(tea.StringValue(hostname.Hostname)) == nil
			})
			if len(hostnames) == 0 {
				return nil, errors.New("could not find any hostnames matched by certificate")
			}

			hostnameIds = lo.Map(hostnames, func(hostname *aliesa.ListCustomHostnamesResponseBodyHostnames, _ int) int64 {
				return tea.Int64Value(hostname.HostnameId)
			})
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(hostnameIds) == 0 {
		d.logger.Info("no esa saas hostnames to deploy")
	} else {
		d.logger.Info("found esa saas hostnames to deploy", slog.Any("hostnameIds", hostnameIds))
		var errs []error

		certIdentifier := upres.ExtendedData["CertIdentifier"].(string)
		certIdentifierSeps := strings.SplitN(certIdentifier, "-", 2)
		if len(certIdentifierSeps) != 2 {
			return nil, fmt.Errorf("received invalid certificate identifier: '%s'", certIdentifier)
		}

		certId, _ := strconv.ParseInt(certIdentifierSeps[0], 10, 64)
		certRegion := certIdentifierSeps[1]
		for _, hostnameId := range hostnameIds {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateHostnameCertificate(ctx, hostnameId, certId, certRegion); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) getAllHostnames(ctx context.Context) ([]*aliesa.ListCustomHostnamesResponseBodyHostnames, error) {
	hostnames := make([]*aliesa.ListCustomHostnamesResponseBodyHostnames, 0)

	// 查询 SaaS 域名列表
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-getcustomhostname
	listCustomHostnamesPageNumber := 1
	listCustomHostnamesPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCustomHostnamesReq := &aliesa.ListCustomHostnamesRequest{
			SiteId:     tea.Int64(d.config.SiteId),
			PageNumber: tea.Int32(int32(listCustomHostnamesPageNumber)),
			PageSize:   tea.Int32(int32(listCustomHostnamesPageSize)),
		}
		listCustomHostnamesResp, err := d.sdkClient.ListCustomHostnamesWithContext(ctx, listCustomHostnamesReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'esa.ListCustomHostnames'", slog.Any("request", listCustomHostnamesReq), slog.Any("response", listCustomHostnamesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'esa.ListCustomHostnames': %w", err)
		}

		if listCustomHostnamesResp.Body == nil {
			break
		}

		ignoredStatuses := []string{"pending", "conflicted", "offline"}
		for _, hostnameItem := range listCustomHostnamesResp.Body.Hostnames {
			if lo.Contains(ignoredStatuses, tea.StringValue(hostnameItem.Status)) {
				continue
			}

			hostnames = append(hostnames, hostnameItem)
		}

		if len(listCustomHostnamesResp.Body.Hostnames) < listCustomHostnamesPageSize {
			break
		}

		listCustomHostnamesPageNumber++
	}

	return hostnames, nil
}

func (d *Deployer) updateHostnameCertificate(ctx context.Context, cloudHostnameId int64, cloudCertId int64, cloudCertRegion string) error {
	// 更新 SaaS 域名
	// REF: https://help.aliyun.com/zh/edge-security-acceleration/esa/api-esa-2024-09-10-updatecustomhostname
	updateCustomHostnameReq := &aliesa.UpdateCustomHostnameRequest{
		HostnameId: tea.Int64(cloudHostnameId),
		SslFlag:    tea.String("on"),
		CertType:   tea.String("cas"),
		CasId:      tea.Int64(cloudCertId),
		CasRegion:  tea.String(cloudCertRegion),
	}
	updateCustomHostnameResp, err := d.sdkClient.UpdateCustomHostnameWithContext(ctx, updateCustomHostnameReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'esa.UpdateCustomHostname'", slog.Any("request", updateCustomHostnameReq), slog.Any("response", updateCustomHostnameResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'esa.UpdateCustomHostname': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.EsaClient, error) {
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

	client, err := internal.NewEsaClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
