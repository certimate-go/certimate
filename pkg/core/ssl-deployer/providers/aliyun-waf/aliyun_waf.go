package aliyunwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	aliwaf "github.com/alibabacloud-go/waf-openapi-20211001/v5/client"

	"github.com/certimate-go/certimate/pkg/core"
	sslmgrsp "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/utils/ifelse"
	xslices "github.com/certimate-go/certimate/pkg/utils/slices"
	xtypes "github.com/certimate-go/certimate/pkg/utils/types"
)

type SSLDeployerProviderConfig struct {
	// 阿里云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 阿里云 AccessKeySecret。
	AccessKeySecret string `json:"accessKeySecret"`
	// 阿里云资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 阿里云地域。
	Region string `json:"region"`
	// 服务版本。
	ServiceVersion string `json:"serviceVersion"`
	// WAF 实例 ID。
	InstanceId string `json:"instanceId"`
	// 接入域名（支持泛域名）。
	Domain string `json:"domain,omitempty"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *slog.Logger
	sdkClient  *aliwaf.Client
	sslManager core.SSLManager
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeySecret, config.Region)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	sslmgr, err := sslmgrsp.NewSSLManagerProvider(&sslmgrsp.SSLManagerProviderConfig{
		AccessKeyId:     config.AccessKeyId,
		AccessKeySecret: config.AccessKeySecret,
		ResourceGroupId: config.ResourceGroupId,
		Region: ifelse.
			If[string](config.Region == "" || strings.HasPrefix(config.Region, "cn-")).
			Then("cn-hangzhou").
			Else("ap-southeast-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sslManager: sslmgr,
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sslManager.SetLogger(logger)
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	if d.config.InstanceId == "" {
		return nil, errors.New("config `instanceId` is required")
	}

	switch d.config.ServiceVersion {
	case "3", "3.0":
		if err := d.deployToWAF3(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported service version '%s'", d.config.ServiceVersion)
	}

	return &core.SSLDeployResult{}, nil
}

func (d *SSLDeployerProvider) deployToWAF3(ctx context.Context, certPEM string, privkeyPEM string) error {
	// 上传证书
	upres, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	if d.config.Domain == "" {
		// 未指定接入域名，只需替换默认证书即可

		// 查询默认 SSL/TLS 设置
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describedefaulthttps
		describeDefaultHttpsReq := &aliwaf.DescribeDefaultHttpsRequest{
			ResourceManagerResourceGroupId: xtypes.ToPtrOrZeroNil(d.config.ResourceGroupId),
			InstanceId:                     tea.String(d.config.InstanceId),
			RegionId:                       tea.String(d.config.Region),
		}
		describeDefaultHttpsResp, err := d.sdkClient.DescribeDefaultHttps(describeDefaultHttpsReq)
		d.logger.Debug("sdk request 'waf.DescribeDefaultHttps'", slog.Any("request", describeDefaultHttpsReq), slog.Any("response", describeDefaultHttpsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.DescribeDefaultHttps': %w", err)
		}

		// 修改默认 SSL/TLS 设置
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifydefaulthttps
		modifyDefaultHttpsReq := &aliwaf.ModifyDefaultHttpsRequest{
			ResourceManagerResourceGroupId: xtypes.ToPtrOrZeroNil(d.config.ResourceGroupId),
			InstanceId:                     tea.String(d.config.InstanceId),
			RegionId:                       tea.String(d.config.Region),
			CertId:                         tea.String(upres.CertId),
			TLSVersion:                     tea.String("tlsv1"),
			EnableTLSv3:                    tea.Bool(false),
		}
		if describeDefaultHttpsResp.Body != nil && describeDefaultHttpsResp.Body.DefaultHttps != nil {
			modifyDefaultHttpsReq.TLSVersion = describeDefaultHttpsResp.Body.DefaultHttps.TLSVersion
			modifyDefaultHttpsReq.EnableTLSv3 = describeDefaultHttpsResp.Body.DefaultHttps.EnableTLSv3
		}
		modifyDefaultHttpsResp, err := d.sdkClient.ModifyDefaultHttps(modifyDefaultHttpsReq)
		d.logger.Debug("sdk request 'waf.ModifyDefaultHttps'", slog.Any("request", modifyDefaultHttpsReq), slog.Any("response", modifyDefaultHttpsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ModifyDefaultHttps': %w", err)
		}
	} else {
		// 指定接入域名

		// 查询 CNAME 接入详情
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describedomaindetail
		describeDomainDetailReq := &aliwaf.DescribeDomainDetailRequest{
			InstanceId: tea.String(d.config.InstanceId),
			RegionId:   tea.String(d.config.Region),
			Domain:     tea.String(d.config.Domain),
		}
		describeDomainDetailResp, err := d.sdkClient.DescribeDomainDetail(describeDomainDetailReq)
		d.logger.Debug("sdk request 'waf.DescribeDomainDetail'", slog.Any("request", describeDomainDetailReq), slog.Any("response", describeDomainDetailResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.DescribeDomainDetail': %w", err)
		}

		// 修改 CNAME 接入资源
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifydomain
		modifyDomainReq := &aliwaf.ModifyDomainRequest{
			InstanceId: tea.String(d.config.InstanceId),
			RegionId:   tea.String(d.config.Region),
			Domain:     tea.String(d.config.Domain),
			Listen:     &aliwaf.ModifyDomainRequestListen{CertId: tea.String(upres.ExtendedData["certIdentifier"].(string))},
			Redirect:   &aliwaf.ModifyDomainRequestRedirect{Loadbalance: tea.String("iphash")},
		}
		modifyDomainReq = assign(modifyDomainReq, describeDomainDetailResp.Body)
		modifyDomainResp, err := d.sdkClient.ModifyDomain(modifyDomainReq)
		d.logger.Debug("sdk request 'waf.ModifyDomain'", slog.Any("request", modifyDomainReq), slog.Any("response", modifyDomainResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ModifyDomain': %w", err)
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*aliwaf.Client, error) {
	// 接入点一览：https://api.aliyun.com/product/waf-openapi
	var endpoint string
	switch region {
	case "":
		endpoint = "wafopenapi.cn-hangzhou.aliyuncs.com"
	default:
		endpoint = fmt.Sprintf("wafopenapi.%s.aliyuncs.com", region)
	}

	config := &aliopen.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := aliwaf.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func assign(source *aliwaf.ModifyDomainRequest, target *aliwaf.DescribeDomainDetailResponseBody) *aliwaf.ModifyDomainRequest {
	// `ModifyDomain` 中不传的字段表示使用默认值、而非保留原值，
	// 因此这里需要把原配置中的参数重新赋值回去。

	if target == nil {
		return source
	}

	if target.Listen != nil {
		if source.Listen == nil {
			source.Listen = &aliwaf.ModifyDomainRequestListen{}
		}

		if target.Listen.CipherSuite != nil {
			source.Listen.CipherSuite = tea.Int32(int32(*target.Listen.CipherSuite))
		}

		if target.Listen.CustomCiphers != nil {
			source.Listen.CustomCiphers = target.Listen.CustomCiphers
		}

		if target.Listen.EnableTLSv3 != nil {
			source.Listen.EnableTLSv3 = target.Listen.EnableTLSv3
		}

		if target.Listen.ExclusiveIp != nil {
			source.Listen.ExclusiveIp = target.Listen.ExclusiveIp
		}

		if target.Listen.FocusHttps != nil {
			source.Listen.FocusHttps = target.Listen.FocusHttps
		}

		if target.Listen.Http2Enabled != nil {
			source.Listen.Http2Enabled = target.Listen.Http2Enabled
		}

		if target.Listen.HttpPorts != nil {
			source.Listen.HttpPorts = xslices.Map(target.Listen.HttpPorts, func(v *int64) *int32 {
				if v == nil {
					return nil
				}
				return tea.Int32(int32(*v))
			})
		}

		if target.Listen.HttpsPorts != nil {
			source.Listen.HttpsPorts = xslices.Map(target.Listen.HttpsPorts, func(v *int64) *int32 {
				if v == nil {
					return nil
				}
				return tea.Int32(int32(*v))
			})
		}

		if target.Listen.IPv6Enabled != nil {
			source.Listen.IPv6Enabled = target.Listen.IPv6Enabled
		}

		if target.Listen.ProtectionResource != nil {
			source.Listen.ProtectionResource = target.Listen.ProtectionResource
		}

		if target.Listen.TLSVersion != nil {
			source.Listen.TLSVersion = target.Listen.TLSVersion
		}

		if target.Listen.XffHeaderMode != nil {
			source.Listen.XffHeaderMode = tea.Int32(int32(*target.Listen.XffHeaderMode))
		}

		if target.Listen.XffHeaders != nil {
			source.Listen.XffHeaders = target.Listen.XffHeaders
		}
	}

	if target.Redirect != nil {
		if source.Redirect == nil {
			source.Redirect = &aliwaf.ModifyDomainRequestRedirect{}
		}

		if target.Redirect.Backends != nil {
			source.Redirect.Backends = xslices.Map(target.Redirect.Backends, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectBackends) *string {
				if v == nil {
					return nil
				}
				return v.Backend
			})
		}

		if target.Redirect.BackupBackends != nil {
			source.Redirect.BackupBackends = xslices.Map(target.Redirect.BackupBackends, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectBackupBackends) *string {
				if v == nil {
					return nil
				}
				return v.Backend
			})
		}

		if target.Redirect.ConnectTimeout != nil {
			source.Redirect.ConnectTimeout = target.Redirect.ConnectTimeout
		}

		if target.Redirect.FocusHttpBackend != nil {
			source.Redirect.FocusHttpBackend = target.Redirect.FocusHttpBackend
		}

		if target.Redirect.Keepalive != nil {
			source.Redirect.Keepalive = target.Redirect.Keepalive
		}

		if target.Redirect.KeepaliveRequests != nil {
			source.Redirect.KeepaliveRequests = target.Redirect.KeepaliveRequests
		}

		if target.Redirect.KeepaliveTimeout != nil {
			source.Redirect.KeepaliveTimeout = target.Redirect.KeepaliveTimeout
		}

		if target.Redirect.Loadbalance != nil {
			source.Redirect.Loadbalance = target.Redirect.Loadbalance
		}

		if target.Redirect.ReadTimeout != nil {
			source.Redirect.ReadTimeout = target.Redirect.ReadTimeout
		}

		if target.Redirect.RequestHeaders != nil {
			source.Redirect.RequestHeaders = xslices.Map(target.Redirect.RequestHeaders, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectRequestHeaders) *aliwaf.ModifyDomainRequestRedirectRequestHeaders {
				if v == nil {
					return nil
				}
				return &aliwaf.ModifyDomainRequestRedirectRequestHeaders{
					Key:   v.Key,
					Value: v.Value,
				}
			})
		}

		if target.Redirect.Retry != nil {
			source.Redirect.Retry = target.Redirect.Retry
		}

		if target.Redirect.SniEnabled != nil {
			source.Redirect.SniEnabled = target.Redirect.SniEnabled
		}

		if target.Redirect.SniHost != nil {
			source.Redirect.SniHost = target.Redirect.SniHost
		}

		if target.Redirect.WriteTimeout != nil {
			source.Redirect.WriteTimeout = target.Redirect.WriteTimeout
		}

		if target.Redirect.XffProto != nil {
			source.Redirect.XffProto = target.Redirect.XffProto
		}
	}

	return source
}
