package volcengineclbdomainextensions

import (
	"context"
	"fmt"
	"log/slog"

	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	veclb "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/volcengine/volcengine-go-sdk/service/clb"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 火山引擎项目名称。
	ProjectName string `json:"projectName,omitempty"`
	// 火山引擎地域。
	Region string `json:"region"`
	// 负载均衡监听器 ID。
	ListenerId string `json:"listenerId"`
	// 要部署证书的扩展域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *veclb.CLB
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:     config.AccessKeyId,
		SecretAccessKey: config.SecretAccessKey,
		ProjectName:     config.ProjectName,
		Region:          config.Region,
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
	if d.config.ListenerId == "" {
		return nil, fmt.Errorf("config `listenerId` is required")
	}
	if d.config.Domain == "" {
		return nil, fmt.Errorf("config `domain` is required")
	}

	// 上传证书到火山引擎证书中心，后续通过证书 ID 绑定到扩展域名。
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	}
	d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))

	if err := d.updateListenerDomainExtension(ctx, upres.CertId); err != nil {
		return nil, err
	}

	return &DeployResult{}, nil
}

func (d *Deployer) updateListenerDomainExtension(ctx context.Context, cloudCertId string) error {
	// 查询监听器详情以取得扩展域名 ID。
	// REF: https://www.volcengine.com/docs/6406/71778
	describeListenerAttributesReq := &veclb.DescribeListenerAttributesInput{
		ListenerId: ve.String(d.config.ListenerId),
	}
	describeListenerAttributesResp, err := d.sdkClient.DescribeListenerAttributesWithContext(ctx, describeListenerAttributesReq)
	d.logger.Debug("sdk request 'clb.DescribeListenerAttributes'", slog.Any("request", describeListenerAttributesReq), slog.Any("response", describeListenerAttributesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.DescribeListenerAttributes': %w", err)
	}
	if describeListenerAttributesResp == nil {
		return fmt.Errorf("could not find clb listener '%s'", d.config.ListenerId)
	}

	var domainExtensionId *string
	for _, domainExtension := range describeListenerAttributesResp.DomainExtensions {
		if domainExtension == nil || ve.StringValue(domainExtension.Domain) != d.config.Domain {
			continue
		}
		domainExtensionId = domainExtension.DomainExtensionId
		break
	}
	if ve.StringValue(domainExtensionId) == "" {
		return fmt.Errorf("could not find clb listener domain extension '%s' for listener '%s'", d.config.Domain, d.config.ListenerId)
	}

	// 修改指定监听器的扩展域名证书。
	// REF: https://www.volcengine.com/docs/6406/2193110
	modifyListenerDomainExtensionsReq := &veclb.ModifyListenerDomainExtensionsInput{
		ListenerId: ve.String(d.config.ListenerId),
		ModifyDomainExtensions: []*veclb.ModifyDomainExtensionForModifyListenerDomainExtensionsInput{
			{
				DomainExtensionId:       domainExtensionId,
				CertificateSource:       ve.String("cert_center"),
				CertCenterCertificateId: ve.String(cloudCertId),
			},
		},
	}
	modifyListenerDomainExtensionsResp, err := d.sdkClient.ModifyListenerDomainExtensionsWithContext(ctx, modifyListenerDomainExtensionsReq)
	d.logger.Debug("sdk request 'clb.ModifyListenerDomainExtensions'", slog.Any("request", modifyListenerDomainExtensionsReq), slog.Any("response", modifyListenerDomainExtensionsResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'clb.ModifyListenerDomainExtensions': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey, region string) (*veclb.CLB, error) {
	config := ve.NewConfig().
		WithAkSk(accessKeyId, secretAccessKey).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := veclb.New(session)
	return client, nil
}
