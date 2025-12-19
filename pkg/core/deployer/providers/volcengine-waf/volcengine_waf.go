package volcenginewaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	vewaf "github.com/volcengine/volcengine-go-sdk/service/waf"
	ve "github.com/volcengine/volcengine-go-sdk/volcengine"
	vesession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-waf/internal"
)

type DeployerConfig struct {
	// 火山引擎 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 火山引擎 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 火山引擎地域。
	Region string `json:"region"`
	// WAF 接入模式。
	AccessMode string `json:"accessMode"`
	// 加速域名（支持泛域名）。
	Domain string `json:"domain"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.WafClient
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据接入方式决定部署方式
	switch d.config.AccessMode {
	case ACCESS_MODE_CNAME:
		if err := d.deployWithCNAME(ctx, upres.CertId); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported access mode '%s'", d.config.AccessMode)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployWithCNAME(ctx context.Context, cloudCertId string) error {
	if d.config.Domain == "" {
		return errors.New("config `domain` is required")
	}

	// 查询云 WAF 实例防护网站信息
	// REF: https://www.volcengine.com/docs/6511/1214827
	listDomainReq := &vewaf.ListDomainInput{
		Region:        ve.String(d.config.Region),
		Domain:        ve.String(d.config.Domain),
		AccurateQuery: ve.Int32(1),
		Page:          ve.Int32(1),
		PageSize:      ve.Int32(1),
	}
	listDomainResp, err := d.sdkClient.ListDomain(listDomainReq)
	d.logger.Debug("sdk request 'waf.ListDomain'", slog.Any("request", listDomainReq), slog.Any("response", listDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.ListDomain': %w", err)
	} else if len(listDomainResp.Data) == 0 {
		return fmt.Errorf("could not find domain '%s'", d.config.Domain)
	}

	// 更新云 WAF 实例的防护网站信息
	// REF: https://www.volcengine.com/docs/6511/1214835
	domainInfo := listDomainResp.Data[0]
	updateDomainReq := &vewaf.UpdateDomainInput{
		Region:     ve.String(d.config.Region),
		Domain:     ve.String(d.config.Domain),
		AccessMode: ve.Int32(10),
		Protocols:  ve.StringSlice([]string{"HTTP", "HTTPS"}),
		ProtocolPorts: &vewaf.ProtocolPortsForUpdateDomainInput{
			HTTP:  ve.Int32Slice([]int32{80}),
			HTTPS: ve.Int32Slice([]int32{443}),
		},
		VolcCertificateID:   ve.String(cloudCertId),
		CertificatePlatform: ve.String("certificate-service"),
	}
	if domainInfo.Protocols != nil {
		protocols := strings.Split(ve.StringValue(domainInfo.Protocols), ",")
		if !lo.Contains(protocols, "HTTPS") {
			protocols = append(protocols, "HTTPS")
		}
		updateDomainReq.Protocols = ve.StringSlice(protocols)
	}
	if domainInfo.ProtocolPorts != nil {
		updateDomainReq.ProtocolPorts.HTTP = domainInfo.ProtocolPorts.HTTP
		if domainInfo.ProtocolPorts.HTTPS != nil {
			updateDomainReq.ProtocolPorts.HTTPS = domainInfo.ProtocolPorts.HTTPS
		}
	}
	updateDomainResp, err := d.sdkClient.UpdateDomain(updateDomainReq)
	d.logger.Debug("sdk request 'waf.UpdateDomain'", slog.Any("request", updateDomainReq), slog.Any("response", updateDomainResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.UpdateDomain': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.WafClient, error) {
	config := ve.NewConfig().
		WithAkSk(accessKeyId, accessKeySecret).
		WithRegion(region)

	session, err := vesession.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := internal.NewWafClient(session)
	return client, nil
}
