package volcenginevod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/volcengine/volc-sdk-golang/service/vod"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/business"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/request"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 空间名称
	SpaceName string `json:"spaceName"`
	// 域名类型 [DOMAIN_TYPE_PLAY] [DOMAIN_TYPE_IMAGE]
	DomainType string `json:"domainType"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 点播域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *vod.Vod
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client := vod.NewInstance()
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

	// 遍历配置证书
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
	listDomainDetailPageNum := int32(1)
	listDomainDetailPageSize := int32(1000)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		listDomainReq := &request.VodListDomainRequest{
			SpaceName:         d.config.SpaceName,
			DomainType:        d.config.DomainType,
			SourceStationType: 1,
			Offset:            (listDomainDetailPageNum - 1) * listDomainDetailPageSize,
			Limit:             listDomainDetailPageSize,
		}
		listDomainResp, _, err := d.sdkClient.ListDomain(listDomainReq)
		d.logger.Debug("sdk request 'vod.ListDomain'", slog.Any("request", listDomainReq), slog.Any("response", listDomainResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'vod.ListDomain': %w", err)
		}

		if listDomainResp.Result == nil {
			break
		}

		var byteInstances []*business.VodDomainInstanceInfo
		if d.config.DomainType == DOMAIN_TYPE_PLAY {
			byteInstances = listDomainResp.GetResult().GetPlayInstanceInfo().GetByteInstances()
		} else if d.config.DomainType == DOMAIN_TYPE_IMAGE {
			byteInstances = listDomainResp.GetResult().GetImageInstanceInfo().GetByteInstances()
		} else {
			return nil, fmt.Errorf("unsupported domain type: '%s'", d.config.DomainType)
		}
		for _, byteDomains := range byteInstances {
			if byteDomains.Domains == nil {
				break
			}
			for _, domainItem := range byteDomains.Domains {
				domains = append(domains, domainItem.Domain)
			}
		}

		if int32(listDomainResp.Result.Offset) < listDomainDetailPageSize {
			break
		}

		listDomainDetailPageNum++
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain string, cloudCertId string) error {
	// 更新域名配置
	// REF: https://www.volcengine.com/docs/4/1317310
	updateCertReq := &request.VodUpdateDomainConfigRequest{
		SpaceName:  d.config.SpaceName,
		DomainType: d.config.DomainType,
		Domain:     domain,
		Config: &business.VodDomainConfig{
			HTTPS: &business.HTTPS{
				Switch: ve.Bool(true),
				CertInfo: &business.CertInfo{
					CertId: &cloudCertId,
				},
			},
		},
	}
	updateCertRes, _, err := d.sdkClient.UpdateDomainConfig(updateCertReq)
	d.logger.Debug("sdk request 'vod.UpdateDomainConfig'", slog.Any("request", updateCertReq), slog.Any("response", updateCertRes))
	if err != nil {
		return err
	}

	return nil
}
