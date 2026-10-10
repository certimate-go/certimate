package aliyunmse

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	alimse "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/mse-20190531/v5/client"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xalibabacloud "github.com/certimate-go/certimate/pkg/utils/third-party/alibabacloud"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	Region          string `json:"region"`
	// 云原生网关唯一标识（GatewayUniqueId）。
	GatewayId string `json:"gatewayId"`
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	Domain             string `json:"domain,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *alimse.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}
	if strings.TrimSpace(config.Region) == "" {
		return nil, fmt.Errorf("config `region` is required")
	}
	if strings.TrimSpace(config.GatewayId) == "" {
		return nil, fmt.Errorf("config `gatewayId` is required")
	}
	switch config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT, DOMAIN_MATCH_PATTERN_WILDCARD:
		if strings.TrimSpace(config.Domain) == "" {
			return nil, fmt.Errorf("config `domain` is required")
		}
	case DOMAIN_MATCH_PATTERN_CERTSAN:
	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", config.DomainMatchPattern)
	}

	client, err := alimse.NewClient(&aliopen.Config{
		AccessKeyId:     tea.String(config.AccessKeyId),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		RegionId:        tea.String(config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}
	certmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region:          lo.Ternary(xalibabacloud.IsIntlRegion(config.Region), "ap-southeast-1", ""),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}
	return &Deployer{config: config, logger: slog.Default(), sdkClient: client, sdkCertmgr: certmgr}, nil
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
	cert, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	candidates, err := d.getAllDomains(ctx)
	if err != nil {
		return nil, err
	}
	domains := make([]*alimse.ListGatewayDomainResponseBodyData, 0)
	for _, domain := range candidates {
		if domain == nil || tea.Int64Value(domain.Id) <= 0 || tea.StringValue(domain.Name) == "*" {
			continue
		}
		protocol := tea.StringValue(domain.Protocol)
		if !strings.EqualFold(protocol, "HTTP") && !strings.EqualFold(protocol, "HTTPS") {
			continue
		}
		name := tea.StringValue(domain.Name)
		covered := xcerthostname.IsMatchByCertificate(cert, name)
		var matched bool
		switch d.config.DomainMatchPattern {
		case "", DOMAIN_MATCH_PATTERN_EXACT:
			matched = strings.EqualFold(strings.TrimSpace(d.config.Domain), name)
		case DOMAIN_MATCH_PATTERN_WILDCARD:
			matched = xcerthostname.IsMatch(strings.TrimSpace(d.config.Domain), name)
		case DOMAIN_MATCH_PATTERN_CERTSAN:
			matched = covered
		}
		if !matched {
			continue
		}
		if !covered {
			return nil, fmt.Errorf("certificate does not cover MSE domain %q (id %d)", name, tea.Int64Value(domain.Id))
		}
		domains = append(domains, domain)
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("could not find any matching domains in MSE gateway %q", d.config.GatewayId)
	}
	d.logger.Info("found mse domains to deploy", slog.Any("domains", domains))

	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	}
	certIdentifier, _ := upres.ExtendedData["CertIdWithRegion"].(string)
	if certIdentifier == "" {
		return nil, fmt.Errorf("missing CertIdWithRegion in certificate upload result")
	}
	d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain *alimse.ListGatewayDomainResponseBodyData, _ int) error {
		if err := d.updateDomainCertificate(ctx, domain, certIdentifier); err != nil {
			return fmt.Errorf("failed to update MSE domain %q (id %d): %w", tea.StringValue(domain.Name), tea.Int64Value(domain.Id), err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &DeployResult{}, nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]*alimse.ListGatewayDomainResponseBodyData, error) {
	// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-listgatewaydomain
	req := &alimse.ListGatewayDomainRequest{GatewayUniqueId: tea.String(d.config.GatewayId)}
	resp, err := d.sdkClient.ListGatewayDomainWithContext(ctx, req, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'mse.ListGatewayDomain'", slog.Any("request", req), slog.Any("response", resp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'mse.ListGatewayDomain': %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("empty response from 'mse.ListGatewayDomain'")
	}
	if !tea.BoolValue(resp.Body.Success) {
		return nil, fmt.Errorf("mse.ListGatewayDomain failed: code=%d, message=%s, requestId=%s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId))
	}
	return resp.Body.Data, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain *alimse.ListGatewayDomainResponseBodyData, certIdentifier string) error {
	if strings.EqualFold(tea.StringValue(domain.Protocol), "HTTPS") {
		if tea.StringValue(domain.CertIdentifier) == certIdentifier {
			d.logger.Info("ssl certificate already deployed", slog.String("domain", tea.StringValue(domain.Name)))
			return nil
		}

		// 仅更换证书，避免覆盖域名的跳转、HTTP/2、TLS 和双向认证配置。
		// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-updatesslcert
		req := &alimse.UpdateSSLCertRequest{
			GatewayUniqueId: tea.String(d.config.GatewayId),
			DomainId:        domain.Id,
			CertIdentifier:  tea.String(certIdentifier),
		}
		resp, err := d.sdkClient.UpdateSSLCertWithContext(ctx, req, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'mse.UpdateSSLCert'", slog.Any("request", req), slog.Any("response", resp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'mse.UpdateSSLCert': %w", err)
		}
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("empty response from 'mse.UpdateSSLCert'")
		}
		if !tea.BoolValue(resp.Body.Success) || !tea.BoolValue(resp.Body.Data) {
			return fmt.Errorf("mse.UpdateSSLCert failed: code=%d, message=%s, requestId=%s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId))
		}
		return nil
	}

	// HTTP 首次启用 HTTPS 使用服务端默认 TLS 配置，不启用强制跳转。
	// REF: https://help.aliyun.com/zh/mse/developer-reference/api-mse-2019-05-31-updategatewaydomain
	req := &alimse.UpdateGatewayDomainRequest{
		GatewayUniqueId: tea.String(d.config.GatewayId),
		Id:              domain.Id,
		Protocol:        tea.String("HTTPS"),
		CertIdentifier:  tea.String(certIdentifier),
		MustHttps:       tea.Bool(false),
	}
	resp, err := d.sdkClient.UpdateGatewayDomainWithContext(ctx, req, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'mse.UpdateGatewayDomain'", slog.Any("request", req), slog.Any("response", resp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'mse.UpdateGatewayDomain': %w", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("empty response from 'mse.UpdateGatewayDomain'")
	}
	if !tea.BoolValue(resp.Body.Success) || tea.Int64Value(resp.Body.Data) != tea.Int64Value(domain.Id) {
		return fmt.Errorf("mse.UpdateGatewayDomain failed: code=%d, message=%s, requestId=%s", tea.Int32Value(resp.Body.Code), tea.StringValue(resp.Body.Message), tea.StringValue(resp.Body.RequestId))
	}
	return nil
}
