package volcenginevod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	vevod "github.com/volcengine/volc-sdk-golang/service/vod"
	vevodbusiness "github.com/volcengine/volc-sdk-golang/service/vod/models/business"
	vevodrequest "github.com/volcengine/volc-sdk-golang/service/vod/models/request"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 点播空间名称。
	SpaceName string `json:"spaceName"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 点播域名类型。
	DomainType string `json:"domainType"`
	// 点播加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *vevod.Vod
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client := vevod.NewInstance()
	client.SetAccessKey(config.AccessKeyId)
	client.SetSecretKey(config.AccessKeySecret)

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			domains = append(domains, d.config.Domain)
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, errors.New("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getAllDomains(ctx)
				if err != nil {
					return nil, err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return nil, errors.New("could not find any domains matched by wildcard")
				}
			} else {
				domains = append(domains, d.config.Domain)
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			certX509, err := xcert.ParseCertificateFromPEM(certPEM)
			if err != nil {
				return nil, err
			}

			domainCandidates, err := d.getAllDomains(ctx)
			if err != nil {
				return nil, err
			}

			domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
				return certX509.VerifyHostname(domain) == nil
			})
			if len(domains) == 0 {
				return nil, errors.New("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 遍历更新域名证书
	if len(domains) == 0 {
		d.logger.Info("no vod domains to deploy")
	} else {
		d.logger.Info("found vod domains to deploy", slog.Any("domains", domains))
		var errs []error

		for _, domain := range domains {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				if err := d.updateDomainCertificate(ctx, domain, upres.CertId); err != nil {
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

func (d *Deployer) getAllDomains(ctx context.Context) ([]string, error) {
	domains := make([]string, 0)

	// 获取空间域名列表
	// REF: https://www.volcengine.com/docs/4/106062
	listDomainDetailOffset := 0
	listDomainDetailLimit := 1000
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listDomainReq := &vevodrequest.VodListDomainRequest{
			SpaceName:         d.config.SpaceName,
			DomainType:        d.config.DomainType,
			SourceStationType: 1,
			Offset:            int32(listDomainDetailOffset),
			Limit:             int32(listDomainDetailLimit),
		}
		listDomainResp, _, err := d.sdkClient.ListDomain(listDomainReq)
		d.logger.Debug("sdk request 'vod.ListDomain'", slog.Any("request", listDomainReq), slog.Any("response", listDomainResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.ListDomain': %w", err)
		}

		if listDomainResp.Result == nil {
			break
		}

		var domainInstances []*vevodbusiness.VodDomainInstanceInfo
		switch d.config.DomainType {
		case DOMAIN_TYPE_PLAY:
			domainInstances = listDomainResp.GetResult().GetPlayInstanceInfo().GetByteInstances()
		case DOMAIN_TYPE_IMAGE:
			domainInstances = listDomainResp.GetResult().GetImageInstanceInfo().GetByteInstances()
		default:
			return nil, fmt.Errorf("unsupported domain type: '%s'", d.config.DomainType)
		}

		for _, domainInstance := range domainInstances {
			if domainInstance.Domains == nil {
				continue
			}
			for _, domainItem := range domainInstance.Domains {
				domains = append(domains, domainItem.Domain)
			}
		}

		if listDomainResp.Result.Total <= int64(listDomainDetailOffset+listDomainDetailLimit) {
			break
		}

		listDomainDetailOffset += listDomainDetailLimit
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 更新域名配置
	// REF: https://www.volcengine.com/docs/4/1317310
	updateDomainConfigReq := &vevodrequest.VodUpdateDomainConfigRequest{
		SpaceName:  d.config.SpaceName,
		DomainType: d.config.DomainType,
		Domain:     domain,
		Config: &vevodbusiness.VodDomainConfig{
			HTTPS: &vevodbusiness.HTTPS{
				Switch: ve.Bool(true),
				CertInfo: &vevodbusiness.CertInfo{
					CertId: &cloudCertId,
				},
			},
		},
	}
	updateDomainConfigResp, _, err := d.sdkClient.UpdateDomainConfig(updateDomainConfigReq)
	d.logger.Debug("sdk request 'vod.UpdateDomainConfig'", slog.Any("request", updateDomainConfigReq), slog.Any("response", updateDomainConfigResp))
	if err != nil {
		return err
	}

	return nil
}
