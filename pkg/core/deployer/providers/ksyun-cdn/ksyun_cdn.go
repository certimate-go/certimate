package ksyuncdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/KscSDK/ksc-sdk-go/ksc"
	ksccdnv1 "github.com/KscSDK/ksc-sdk-go/service/cdnv1"
	"github.com/go-viper/mapstructure/v2"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 金山云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 金山云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ksccdnv1.Cdnv1
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
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
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_DOMAIN:
		if err := d.deployToDomain(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToDomain(ctx context.Context, certPEM, privkeyPEM string) error {
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
				domainCandidates, err := d.getAllDomains(ctx)
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
			domainCandidates, err := d.getAllDomains(ctx)
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
		d.logger.Info("no cdn domains to deploy")
	} else {
		d.logger.Info("found cdn domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, certPEM, privkeyPEM); err != nil {
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

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 更新证书
	// https://docs.ksyun.com/documents/259
	setCertificateInput := map[string]any{
		"CertificateId":     d.config.CertificateId,
		"CertificateName":   fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
		"ServerCertificate": certPEM,
		"PrivateKey":        privkeyPEM,
	}
	setCertificateOutput, err := d.sdkClient.SetCertificatePostWithContext(ctx, &setCertificateInput)
	d.logger.Debug("sdk request 'cdn.SetCertificate'", slog.Any("request", setCertificateInput), slog.Any("response", setCertificateOutput))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.SetCertificate': %w", err)
	}

	return nil
}

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 查询域名列表
	// https://docs.ksyun.com/documents/198
	getCdnDomainsPageNumber := 1
	getCdnDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		getCdnDomainsInput := map[string]any{
			"PageNumber": getCdnDomainsPageNumber,
			"PageSize":   getCdnDomainsPageSize,
		}
		getCdnDomainsOutput, err := d.sdkClient.GetCdnDomainsPostWithContext(ctx, &getCdnDomainsInput)
		d.logger.Debug("sdk request 'cdn.GetCdnDomains'", slog.Any("request", getCdnDomainsInput), slog.Any("response", getCdnDomainsOutput))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'cdn.GetCdnDomains': %w", err)
		}

		type GetCdnDomainsResponse struct {
			PageNumber int32 `json:"PageNumber"`
			PageSize   int32 `json:"PageSize"`
			TotalCount int32 `json:"TotalCount"`
			Domains    []*struct {
				DomainId     string `json:"DomainId"`
				DomainName   string `json:"DomainName"`
				Cname        string `json:"Cname"`
				CdnType      string `json:"CdnType"`
				CreatedTime  string `json:"CreatedTime"`
				ModifiedTime string `json:"ModifiedTime"`
				Region       string `json:"Region"`
			} `json:"Domains"`
		}
		var getCdnDomainsResp *GetCdnDomainsResponse
		mapstructure.Decode(getCdnDomainsOutput, &getCdnDomainsResp)
		if getCdnDomainsResp == nil {
			break
		}

		for _, domainItem := range getCdnDomainsResp.Domains {
			domains = append(domains, domainItem.DomainName)
		}

		if len(getCdnDomainsResp.Domains) < getCdnDomainsPageSize {
			break
		}

		getCdnDomainsPageNumber++
	}

	return domains, nil
}

func (d *Deployer) findDomainIdByDomain(ctx context.Context, domain string) (string, error) {
	// 查询域名列表
	// https://docs.ksyun.com/documents/198
	getCdnDomainsPageNumber := 1
	getCdnDomainsPageSize := 100
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		getCdnDomainsInput := map[string]any{
			"PageNumber": getCdnDomainsPageNumber,
			"PageSize":   getCdnDomainsPageSize,
			"DomainName": domain,
			"FuzzyMatch": "off",
		}
		getCdnDomainsOutput, err := d.sdkClient.GetCdnDomainsPostWithContext(ctx, &getCdnDomainsInput)
		d.logger.Debug("sdk request 'cdn.GetCdnDomains'", slog.Any("request", getCdnDomainsInput), slog.Any("response", getCdnDomainsOutput))
		if err != nil {
			return "", fmt.Errorf("failed to execute sdk request 'cdn.GetCdnDomains': %w", err)
		}

		type GetCdnDomainsResponse struct {
			PageNumber int32 `json:"PageNumber"`
			PageSize   int32 `json:"PageSize"`
			TotalCount int32 `json:"TotalCount"`
			Domains    []*struct {
				DomainId     string `json:"DomainId"`
				DomainName   string `json:"DomainName"`
				Cname        string `json:"Cname"`
				CdnType      string `json:"CdnType"`
				CreatedTime  string `json:"CreatedTime"`
				ModifiedTime string `json:"ModifiedTime"`
				Region       string `json:"Region"`
			} `json:"Domains"`
		}
		var getCdnDomainsResp *GetCdnDomainsResponse
		mapstructure.Decode(getCdnDomainsOutput, &getCdnDomainsResp)
		if getCdnDomainsResp == nil {
			break
		}

		for _, domainItem := range getCdnDomainsResp.Domains {
			if strings.EqualFold(domainItem.DomainName, domain) {
				return domainItem.DomainId, nil
			}
		}

		if len(getCdnDomainsResp.Domains) < getCdnDomainsPageSize {
			break
		}

		getCdnDomainsPageNumber++
	}

	return "", fmt.Errorf("could not find domain '%s'", domain)
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, certPEM, privkeyPEM string) error {
	// 获取域名 ID
	domainId, err := d.findDomainIdByDomain(ctx, domain)
	if err != nil {
		return err
	}

	// 为加速域名配置证书接口
	// https://docs.ksyun.com/documents/261
	configCertificateInput := map[string]any{
		"Enable":            "on",
		"DomainIds":         domainId,
		"CertificateName":   fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
		"ServerCertificate": certPEM,
		"PrivateKey":        privkeyPEM,
	}
	configCertificateOutput, err := d.sdkClient.ConfigCertificatePostWithContext(ctx, &configCertificateInput)
	d.logger.Debug("sdk request 'cdn.ConfigCertificate'", slog.Any("request", configCertificateInput), slog.Any("response", configCertificateOutput))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'cdn.ConfigCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ksccdnv1.Cdnv1, error) {
	region := "cn-beijing-6"
	client := ksccdnv1.SdkNew(ksc.NewClient(accessKeyId, secretAccessKey), &ksc.Config{Region: &region})
	return client, nil
}
