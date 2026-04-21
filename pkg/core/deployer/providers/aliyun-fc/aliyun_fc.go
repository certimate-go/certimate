package aliyunfc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alifc3 "github.com/alibabacloud-go/fc-20230330/v4/client"
	alifc2 "github.com/alibabacloud-go/fc-open-20210406/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-fc/internal"
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
	// 服务版本。
	// 可取值 "2.0"、"3.0"。
	ServiceVersion string `json:"serviceVersion"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 自定义域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClients *wSDKClients
}

var _ deployer.Provider = (*Deployer)(nil)

type wSDKClients struct {
	FC2 *internal.FcopenClient
	FC3 *internal.FcClient
}

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	clients, err := createSDKClients(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClients: clients,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	switch d.config.ServiceVersion {
	case "3", "3.0":
		if err := d.deployToFC3(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case "2", "2.0":
		if err := d.deployToFC2(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported service version '%s'", d.config.ServiceVersion)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToFC3(ctx context.Context, certPEM, privkeyPEM string) error {
	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getFC3AllDomains(ctx)
				if err != nil {
					return err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getFC3AllDomains(ctx)
			if err != nil {
				return err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no fc domains to deploy")
	} else {
		d.logger.Info("found fc domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateFC3DomainCertificate(ctx, domain, certPEM, privkeyPEM); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) deployToFC2(ctx context.Context, certPEM, privkeyPEM string) error {
	// 获取待部署的域名列表
	var domains []string
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getFC2AllDomains(ctx)
				if err != nil {
					return err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getFC2AllDomains(ctx)
			if err != nil {
				return err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return xcerthostname.IsMatchByCertificatePEM(certPEM, domain)
			})
			if len(domains) == 0 {
				return errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no fc domains to deploy")
	} else {
		d.logger.Info("found fc domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateFC2DomainCertificate(ctx, domain, certPEM, privkeyPEM); err != nil {
					errs = append(errs, err)
				}
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}
	}

	return nil
}

func (d *Deployer) getFC3AllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 列出自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc/developer-reference/api-fc-2023-03-30-listcustomdomains
	listCustomDomainsNextToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCustomDomainsReq := &alifc3.ListCustomDomainsRequest{
			NextToken: listCustomDomainsNextToken,
			Limit:     tea.Int32(100),
		}
		listCustomDomainsResp, err := d.sdkClients.FC3.ListCustomDomainsWithContext(ctx, listCustomDomainsReq, make(map[string]*string, 0), &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'fc.ListCustomDomains'", slog.Any("request", listCustomDomainsReq), slog.Any("response", listCustomDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'fc.ListCustomDomains': %w", err)
		}

		if listCustomDomainsResp.Body == nil {
			break
		}

		for _, domainItem := range listCustomDomainsResp.Body.CustomDomains {
			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}

		if len(listCustomDomainsResp.Body.CustomDomains) == 0 || listCustomDomainsResp.Body.NextToken == nil {
			break
		}

		listCustomDomainsNextToken = listCustomDomainsResp.Body.NextToken
	}

	return domains, nil
}

func (d *Deployer) getFC2AllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 列出自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc-2-0/developer-reference/api-fc-open-2021-04-06-listcustomdomains
	listCustomDomainsNextToken := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listCustomDomainsReq := &alifc2.ListCustomDomainsRequest{
			NextToken: listCustomDomainsNextToken,
			Limit:     tea.Int32(100),
		}
		listCustomDomainsResp, err := d.sdkClients.FC2.ListCustomDomains(listCustomDomainsReq)
		d.logger.Debug("sdk request 'fc.ListCustomDomains'", slog.Any("request", listCustomDomainsReq), slog.Any("response", listCustomDomainsResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'fc.ListCustomDomains': %w", err)
		}

		if listCustomDomainsResp.Body == nil {
			break
		}

		for _, domainItem := range listCustomDomainsResp.Body.CustomDomains {
			domains = append(domains, tea.StringValue(domainItem.DomainName))
		}

		if len(listCustomDomainsResp.Body.CustomDomains) == 0 || listCustomDomainsResp.Body.NextToken == nil {
			break
		}

		listCustomDomainsNextToken = listCustomDomainsResp.Body.NextToken
	}

	return domains, nil
}

func (d *Deployer) updateFC3DomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 获取自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc-3-0/developer-reference/api-fc-2023-03-30-getcustomdomain
	getCustomDomainResp, err := d.sdkClients.FC3.GetCustomDomainWithContext(ctx, tea.String(domain), make(map[string]*string), &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'fc.GetCustomDomain'", slog.Any("response", getCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'fc.GetCustomDomain': %w", err)
	} else {
		if getCustomDomainResp.Body.CertConfig != nil && tea.StringValue(getCustomDomainResp.Body.CertConfig.Certificate) == certPEM {
			return nil
		}
	}

	// 更新自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc-3-0/developer-reference/api-fc-2023-03-30-updatecustomdomain
	updateCustomDomainReq := &alifc3.UpdateCustomDomainRequest{
		Body: &alifc3.UpdateCustomDomainInput{
			CertConfig: &alifc3.CertConfig{
				CertName:    tea.String(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
				Certificate: tea.String(certPEM),
				PrivateKey:  tea.String(privkeyPEM),
			},
			Protocol:  getCustomDomainResp.Body.Protocol,
			TlsConfig: getCustomDomainResp.Body.TlsConfig,
		},
	}
	if tea.StringValue(updateCustomDomainReq.Body.Protocol) == "HTTP" {
		updateCustomDomainReq.Body.Protocol = tea.String("HTTP,HTTPS")
	}
	updateCustomDomainResp, err := d.sdkClients.FC3.UpdateCustomDomainWithContext(ctx, tea.String(domain), updateCustomDomainReq, make(map[string]*string), &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'fc.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'fc.UpdateCustomDomain': %w", err)
	}

	return nil
}

func (d *Deployer) updateFC2DomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 获取自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc-2-0/developer-reference/api-fc-open-2021-04-06-getcustomdomain
	getCustomDomainResp, err := d.sdkClients.FC2.GetCustomDomain(tea.String(domain))
	d.logger.Debug("sdk request 'fc.GetCustomDomain'", slog.Any("response", getCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'fc.GetCustomDomain': %w", err)
	} else {
		if getCustomDomainResp.Body.CertConfig != nil && tea.StringValue(getCustomDomainResp.Body.CertConfig.Certificate) == certPEM {
			return nil
		}
	}

	// 更新自定义域名
	// REF: https://help.aliyun.com/zh/functioncompute/fc-2-0/developer-reference/api-fc-open-2021-04-06-updatecustomdomain
	updateCustomDomainReq := &alifc2.UpdateCustomDomainRequest{
		CertConfig: &alifc2.CertConfig{
			CertName:    tea.String(fmt.Sprintf("certimate-%d", time.Now().UnixMilli())),
			Certificate: tea.String(certPEM),
			PrivateKey:  tea.String(privkeyPEM),
		},
		Protocol:  getCustomDomainResp.Body.Protocol,
		TlsConfig: getCustomDomainResp.Body.TlsConfig,
	}
	if tea.StringValue(updateCustomDomainReq.Protocol) == "HTTP" {
		updateCustomDomainReq.Protocol = tea.String("HTTP,HTTPS")
	}
	updateCustomDomainResp, err := d.sdkClients.FC2.UpdateCustomDomain(tea.String(domain), updateCustomDomainReq)
	d.logger.Debug("sdk request 'fc.UpdateCustomDomain'", slog.Any("request", updateCustomDomainReq), slog.Any("response", updateCustomDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'fc.UpdateCustomDomain': %w", err)
	}

	return nil
}

func createSDKClients(accessKeyId, accessKeySecret, region string) (*wSDKClients, error) {
	// 接入点一览 https://api.aliyun.com/product/FC-Open
	var fc2Endpoint string
	switch region {
	case "":
		fc2Endpoint = "fc.aliyuncs.com"
	case "cn-hangzhou-finance":
		fc2Endpoint = fmt.Sprintf("%s.fc.aliyuncs.com", region)
	default:
		fc2Endpoint = fmt.Sprintf("fc.%s.aliyuncs.com", region)
	}

	fc2Config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(fc2Endpoint),
	}
	fc2Client, err := internal.NewFcopenClient(fc2Config)
	if err != nil {
		return nil, err
	}

	// 接入点一览 https://api.aliyun.com/product/FC
	var fc3Endpoint string
	switch region {
	case "":
		fc3Endpoint = "fcv3.cn-hangzhou.aliyuncs.com"
	case "me-central-1", "cn-hangzhou-finance", "cn-shanghai-finance-1", "cn-heyuan-acdr-1":
		fc3Endpoint = fmt.Sprintf("%s.fc.aliyuncs.com", region)
	default:
		fc3Endpoint = fmt.Sprintf("fcv3.%s.aliyuncs.com", region)
	}

	fc3Config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(fc3Endpoint),
	}
	fc3Client, err := internal.NewFcClient(fc3Config)
	if err != nil {
		return nil, err
	}

	return &wSDKClients{
		FC2: fc2Client,
		FC3: fc3Client,
	}, nil
}
