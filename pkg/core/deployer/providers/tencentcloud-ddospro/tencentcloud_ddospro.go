package tencentcloudddospro

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	tcantiddos "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/antiddos/v20200309"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/tencentcloud-ssl"
	xcerthostname "github.com/certimate-go/certimate/pkg/utils/cert/hostname"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
	xtencentcloud "github.com/certimate-go/certimate/pkg/utils/third-party/tencentcloud"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云项目 ID。
	ProjectId int64 `json:"projectId,omitempty"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// DDoS 高防实例 ID。
	InstanceId string `json:"instanceId"`
	// 域名匹配模式。
	// 零值时默认值 [DOMAIN_MATCH_PATTERN_EXACT]。
	DomainMatchPattern string `json:"domainMatchPattern,omitempty"`
	// 网站域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *tcantiddos.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		ProjectId: config.ProjectId,
		Endpoint:  lo.Ternary(xtencentcloud.IsIntlAPIEndpoint(config.Endpoint), "ssl.intl.tencentcloudapi.com", ""),
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
	if d.config.InstanceId == "" {
		return nil, fmt.Errorf("config `instanceId` is required")
	}
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取待部署的域名列表
	domains := make([]string, 0)
	switch d.config.DomainMatchPattern {
	case "", DOMAIN_MATCH_PATTERN_EXACT:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			domains = []string{d.config.Domain}
		}

	case DOMAIN_MATCH_PATTERN_WILDCARD:
		{
			if d.config.Domain == "" {
				return nil, fmt.Errorf("config `domain` is required")
			}

			if strings.HasPrefix(d.config.Domain, "*.") {
				domainCandidates, err := d.getMatchedDomainsByCertId(ctx, upres.CertId)
				if err != nil {
					return nil, err
				}

				domains = lo.Filter(domainCandidates, func(domain string, _ int) bool {
					return xcerthostname.IsMatch(d.config.Domain, domain)
				})
				if len(domains) == 0 {
					return nil, fmt.Errorf("could not find any domains matched by wildcard")
				}
			} else {
				domains = []string{d.config.Domain}
			}
		}

	case DOMAIN_MATCH_PATTERN_CERTSAN:
		{
			domainCandidates, err := d.getMatchedDomainsByCertId(ctx, upres.CertId)
			if err != nil {
				return nil, err
			}

			domains = domainCandidates
			if len(domains) == 0 {
				return nil, fmt.Errorf("could not find any domains matched by certificate")
			}
		}

	default:
		return nil, fmt.Errorf("unsupported domain match pattern: '%s'", d.config.DomainMatchPattern)
	}

	// 批量绑定证书
	if len(domains) == 0 {
		d.logger.Info("no anti-ddos domains to deploy")
	} else {
		d.logger.Info("found anti-ddos domains to deploy", slog.Any("domains", domains))

		if err := xloop.ForRangeAllWithContext(ctx, domains, func(ctx context.Context, domain string, _ int) error {
			return d.updateDomainCertificate(ctx, domain, upres.CertId)
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func (d *Deployer) getMatchedDomainsByCertId(ctx context.Context, cloudCertId string) ([]string, error) {
	domains := make([]string, 0)

	// 获取指定证书可关联的域名
	// REF: https://cloud.tencent.com/document/api/297/95348
	describeL7RulesBySSLCertIdReq := tcantiddos.NewDescribeL7RulesBySSLCertIdRequest()
	describeL7RulesBySSLCertIdReq.Status = common.StringPtr("all")
	describeL7RulesBySSLCertIdReq.CertIds = []*string{common.StringPtr(cloudCertId)}
	describeL7RulesBySSLCertIdResp, err := d.sdkClient.DescribeL7RulesBySSLCertIdWithContext(ctx, describeL7RulesBySSLCertIdReq)
	d.logger.Debug("sdk request 'antiddos.DescribeL7RulesBySSLCertId'", slog.Any("request", describeL7RulesBySSLCertIdReq), slog.Any("response", describeL7RulesBySSLCertIdResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'antiddos.DescribeL7RulesBySSLCertId': %w", err)
	} else if len(describeL7RulesBySSLCertIdResp.Response.CertSet) == 0 {
		return nil, fmt.Errorf("could not find any domains matched by certificate")
	}

	certEntry := describeL7RulesBySSLCertIdResp.Response.CertSet[0]
	for i := range certEntry.L7Rules {
		if lo.FromPtr(certEntry.L7Rules[i].InsId) == d.config.InstanceId {
			domains = append(domains, lo.FromPtr(certEntry.L7Rules[i].Domain))
		}
	}

	return domains, nil
}

func (d *Deployer) updateDomainCertificate(ctx context.Context, domain, cloudCertId string) error {
	// 高防 IP 获取 7 层规则
	// REF: https://cloud.tencent.com/document/api/297/95344
	var ruleEntry *tcantiddos.NewL7RuleEntry
	describeNewL7RulesOffset := 0
	describeNewL7RulesLimit := 100
	for {
		describeNewL7RulesReq := tcantiddos.NewDescribeNewL7RulesRequest()
		describeNewL7RulesReq.Business = common.StringPtr("bgpip")
		describeNewL7RulesReq.Domain = common.StringPtr(domain)
		describeNewL7RulesReq.Offset = common.Uint64Ptr(uint64(describeNewL7RulesOffset))
		describeNewL7RulesReq.Limit = common.Uint64Ptr(uint64(describeNewL7RulesLimit))
		describeNewL7RulesResp, err := d.sdkClient.DescribeNewL7RulesWithContext(ctx, describeNewL7RulesReq)
		d.logger.Debug("sdk request 'antiddos.DescribeNewL7Rules'", slog.Any("request", describeNewL7RulesReq), slog.Any("response", describeNewL7RulesResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'antiddos.DescribeNewL7Rules': %w", err)
		}

		for _, ruleItem := range describeNewL7RulesResp.Response.Rules {
			if lo.FromPtr(ruleItem.Domain) == domain {
				ruleEntry = ruleItem
				break
			}
		}

		if len(describeNewL7RulesResp.Response.Rules) < describeNewL7RulesLimit {
			break
		}

		describeNewL7RulesOffset += describeNewL7RulesLimit
	}
	if ruleEntry == nil {
		return fmt.Errorf("could not find domain '%s'", domain)
	} else if lo.FromPtr(ruleEntry.CertType) == 2 && lo.FromPtr(ruleEntry.SSLId) == cloudCertId {
		return nil
	}

	// 修改 7 层转发规则
	modifyNewDomainRulesReq := tcantiddos.NewModifyNewDomainRulesRequest()
	modifyNewDomainRulesReq.Business = common.StringPtr("bgpip")
	modifyNewDomainRulesReq.Id = common.StringPtr(d.config.InstanceId)
	modifyNewDomainRulesReq.Rule = &tcantiddos.NewL7RuleEntry{
		Protocol:          ruleEntry.Protocol,
		LbType:            ruleEntry.LbType,
		Domain:            ruleEntry.Domain,
		KeepEnable:        ruleEntry.KeepEnable,
		KeepTime:          ruleEntry.KeepTime,
		SourceType:        ruleEntry.SourceType,
		SourceList:        ruleEntry.SourceList,
		Region:            ruleEntry.Region,
		Id:                ruleEntry.Id,
		Ip:                ruleEntry.Ip,
		RuleId:            ruleEntry.RuleId,
		RuleName:          ruleEntry.RuleName,
		CertType:          common.Uint64Ptr(2),
		SSLId:             common.StringPtr(cloudCertId),
		CCStatus:          ruleEntry.CCStatus,
		CCEnable:          ruleEntry.CCEnable,
		CCThresholdNew:    ruleEntry.CCThresholdNew,
		CCLevel:           ruleEntry.CCLevel,
		HttpsToHttpEnable: ruleEntry.HttpsToHttpEnable,
		VirtualPort:       ruleEntry.VirtualPort,
		RewriteHttps:      ruleEntry.RewriteHttps,
	}
	modifyNewDomainRulesResp, err := d.sdkClient.ModifyNewDomainRulesWithContext(ctx, modifyNewDomainRulesReq)
	d.logger.Debug("sdk request 'antiddos.ModifyNewDomainRules'", slog.Any("request", modifyNewDomainRulesReq), slog.Any("response", modifyNewDomainRulesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'antiddos.ModifyNewDomainRules': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcantiddos.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcantiddos.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
