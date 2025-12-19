package aliyunwaf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	aliwaf "github.com/alibabacloud-go/waf-openapi-20211001/v7/client"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-waf/internal"
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
	// 可取值 "3.0"。
	ServiceVersion string `json:"serviceVersion"`
	// 服务类型。
	ServiceType string `json:"serviceType"`
	// WAF 实例 ID。
	InstanceId string `json:"instanceId"`
	// 云产品类型。
	// 服务类型为 [SERVICE_TYPE_CLOUDRESOURCE] 时必填。
	ResourceProduct string `json:"resourceProduct,omitempty"`
	// 云产品资源 ID。
	// 服务类型为 [SERVICE_TYPE_CLOUDRESOURCE] 时必填。
	ResourceId string `json:"resourceId,omitempty"`
	// 云产品资源端口。
	// 服务类型为 [SERVICE_TYPE_CLOUDRESOURCE] 时必填。
	ResourcePort int32 `json:"resourcePort,omitempty"`
	// 扩展域名（支持泛域名）。
	Domain string `json:"domain,omitempty"`
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
	switch d.config.ServiceVersion {
	case "3", "3.0":
		if err := d.deployToWAF3(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported service version '%s'", d.config.ServiceVersion)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToWAF3(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.InstanceId == "" {
		return errors.New("config `instanceId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 根据接入方式决定部署方式
	switch d.config.ServiceType {
	case SERVICE_TYPE_CLOUDRESOURCE:
		certId := upres.ExtendedData["CertIdentifier"].(string)
		if err := d.deployToWAF3WithCloudResource(ctx, certId); err != nil {
			return err
		}

	case SERVICE_TYPE_CNAME:
		certId := upres.ExtendedData["CertIdentifier"].(string)
		if err := d.deployToWAF3WithCNAME(ctx, certId); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported service version '%s'", d.config.ServiceVersion)
	}

	return nil
}

func (d *Deployer) deployToWAF3WithCloudResource(ctx context.Context, cloudCertId string) error {
	if d.config.ResourceProduct == "" {
		return errors.New("config `resourceProduct` is required")
	}
	if d.config.ResourceId == "" {
		return errors.New("config `resourceId` is required")
	}
	if d.config.ResourcePort == 0 {
		d.config.ResourcePort = 443
	}

	// 查询已同步的云产品资产
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describeproductinstances
	var resourceInstance *aliwaf.DescribeProductInstancesResponseBodyProductInstances
	var resourceInstancePort *aliwaf.DescribeProductInstancesResponseBodyProductInstancesResourcePorts
	describeProductInstancesReq := &aliwaf.DescribeProductInstancesRequest{
		ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
		RegionId:                       tea.String(d.config.Region),
		InstanceId:                     tea.String(d.config.InstanceId),
		ResourceProduct:                tea.String(d.config.ResourceProduct),
		ResourceInstanceId:             tea.String(d.config.ResourceId),
	}
	describeProductInstancesResp, err := d.sdkClient.DescribeProductInstancesWithContext(ctx, describeProductInstancesReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'waf.DescribeProductInstances'", slog.Any("request", describeProductInstancesReq), slog.Any("response", describeProductInstancesResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.DescribeProductInstances': %w", err)
	} else if len(describeProductInstancesResp.Body.ProductInstances) == 0 {
		return fmt.Errorf("could not find waf '%s' cloud resource '%s %s'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId)
	} else {
		resourceInstance = describeProductInstancesResp.Body.ProductInstances[0]

		resourceInstancePort, _ = lo.Find(resourceInstance.ResourcePorts, func(p *aliwaf.DescribeProductInstancesResponseBodyProductInstancesResourcePorts) bool {
			return tea.Int32Value(p.Port) == d.config.ResourcePort
		})
		if resourceInstancePort == nil {
			return fmt.Errorf("could not find waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
		}
	}

	// 查询云产品实例的证书列表
	var resourceInstanceCertificates []*aliwaf.DescribeResourceInstanceCertsResponseBodyCerts = make([]*aliwaf.DescribeResourceInstanceCertsResponseBodyCerts, 0)
	describeResourceInstanceCertsPageNumber := 1
	describeResourceInstanceCertsPageSize := 10
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		describeResourceInstanceCertsReq := &aliwaf.DescribeResourceInstanceCertsRequest{
			ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			InstanceId:                     tea.String(d.config.InstanceId),
			ResourceInstanceId:             tea.String(d.config.ResourceId),
			PageNumber:                     tea.Int64(int64(describeResourceInstanceCertsPageNumber)),
			PageSize:                       tea.Int64(int64(describeResourceInstanceCertsPageSize)),
		}
		describeResourceInstanceCertsResp, err := d.sdkClient.DescribeResourceInstanceCertsWithContext(ctx, describeResourceInstanceCertsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'waf.DescribeResourceInstanceCerts'", slog.Any("request", describeResourceInstanceCertsReq), slog.Any("response", describeResourceInstanceCertsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'wafw.DescribeResourceInstanceCerts': %w", err)
		}

		if describeResourceInstanceCertsResp.Body == nil {
			break
		}

		resourceInstanceCertificates = append(resourceInstanceCertificates, describeResourceInstanceCertsResp.Body.Certs...)

		if len(describeResourceInstanceCertsResp.Body.Certs) < describeResourceInstanceCertsPageSize {
			break
		}

		describeResourceInstanceCertsPageNumber++
	}

	// 生成请求参数
	modifyCloudResourceReq := &aliwaf.ModifyCloudResourceRequest{
		ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
		RegionId:                       tea.String(d.config.Region),
		Listen: &aliwaf.ModifyCloudResourceRequestListen{
			ResourceProduct:    resourceInstance.ResourceProduct,
			ResourceInstanceId: resourceInstance.ResourceInstanceId,
			Protocol:           tea.String("https"),
			Port:               resourceInstancePort.Port,
			Certificates: lo.Map(resourceInstancePort.Certificates, func(c *aliwaf.DescribeProductInstancesResponseBodyProductInstancesResourcePortsCertificates, _ int) *aliwaf.ModifyCloudResourceRequestListenCertificates {
				return &aliwaf.ModifyCloudResourceRequestListenCertificates{
					CertificateId: c.CertificateId,
					AppliedType:   c.AppliedType,
				}
			}),
		},
	}
	if d.config.Domain == "" {
		// 未指定扩展域名，只需替换默认证书
		const certAppliedTypeDefault = "default"
		for _, certItem := range modifyCloudResourceReq.Listen.Certificates {
			if tea.StringValue(certItem.AppliedType) == certAppliedTypeDefault &&
				tea.StringValue(certItem.CertificateId) == cloudCertId {
				return nil
			}
		}

		modifyCloudResourceReq.Listen.Certificates = lo.Filter(modifyCloudResourceReq.Listen.Certificates, func(c *aliwaf.ModifyCloudResourceRequestListenCertificates, _ int) bool {
			return tea.StringValue(c.AppliedType) != certAppliedTypeDefault
		})
		modifyCloudResourceReq.Listen.Certificates = append(modifyCloudResourceReq.Listen.Certificates, &aliwaf.ModifyCloudResourceRequestListenCertificates{
			CertificateId: tea.String(cloudCertId),
			AppliedType:   tea.String(certAppliedTypeDefault),
		})
	} else {
		// 指定扩展域名，需替换扩展证书
		const certAppliedTypeExtension = "extension"

		modifyCloudResourceReq.Listen.Certificates = append(modifyCloudResourceReq.Listen.Certificates, &aliwaf.ModifyCloudResourceRequestListenCertificates{
			CertificateId: tea.String(cloudCertId),
			AppliedType:   tea.String(certAppliedTypeExtension),
		})
	}

	// 过滤掉不存在或已过期的证书，防止接口报错
	modifyCloudResourceReq.Listen.Certificates = lo.Filter(modifyCloudResourceReq.Listen.Certificates, func(c *aliwaf.ModifyCloudResourceRequestListenCertificates, _ int) bool {
		if tea.StringValue(c.CertificateId) == cloudCertId {
			return true
		}

		resourceInstanceCert, _ := lo.Find(resourceInstanceCertificates, func(r *aliwaf.DescribeResourceInstanceCertsResponseBodyCerts) bool {
			cId := tea.StringValue(c.CertificateId)
			rId := tea.StringValue(r.CertIdentifier)
			return cId == rId || strings.Split(cId, "-")[0] == strings.Split(rId, "-")[0]
		})
		if resourceInstanceCert != nil {
			certNotAfter := time.Unix(tea.Int64Value(resourceInstanceCert.AfterDate)/1000, 0)
			return certNotAfter.After(time.Now())
		}

		return false
	})

	// 修改云产品接入的配置
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifycloudresource
	modifyCloudResourceResp, err := d.sdkClient.ModifyCloudResourceWithContext(ctx, modifyCloudResourceReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'waf.ModifyCloudResource'", slog.Any("request", modifyCloudResourceReq), slog.Any("response", modifyCloudResourceResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.ModifyCloudResource': %w", err)
	}

	return nil
}

func (d *Deployer) deployToWAF3WithCNAME(ctx context.Context, cloudCertId string) error {
	if d.config.Domain == "" {
		// 未指定扩展域名，只需替换默认证书

		// 查询默认 SSL/TLS 设置
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describedefaulthttps
		describeDefaultHttpsReq := &aliwaf.DescribeDefaultHttpsRequest{
			ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			RegionId:                       tea.String(d.config.Region),
			InstanceId:                     tea.String(d.config.InstanceId),
		}
		describeDefaultHttpsResp, err := d.sdkClient.DescribeDefaultHttpsWithContext(ctx, describeDefaultHttpsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'waf.DescribeDefaultHttps'", slog.Any("request", describeDefaultHttpsReq), slog.Any("response", describeDefaultHttpsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.DescribeDefaultHttps': %w", err)
		}

		// 修改默认 SSL/TLS 设置
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifydefaulthttps
		modifyDefaultHttpsReq := &aliwaf.ModifyDefaultHttpsRequest{
			ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
			RegionId:                       tea.String(d.config.Region),
			InstanceId:                     tea.String(d.config.InstanceId),
			CertId:                         tea.String(cloudCertId),
			TLSVersion:                     tea.String("tlsv1.2"),
			EnableTLSv3:                    tea.Bool(true),
		}
		if describeDefaultHttpsResp.Body != nil && describeDefaultHttpsResp.Body.DefaultHttps != nil {
			if describeDefaultHttpsResp.Body.DefaultHttps.TLSVersion != nil {
				modifyDefaultHttpsReq.TLSVersion = describeDefaultHttpsResp.Body.DefaultHttps.TLSVersion
			}
			if describeDefaultHttpsResp.Body.DefaultHttps.EnableTLSv3 == nil {
				modifyDefaultHttpsReq.EnableTLSv3 = describeDefaultHttpsResp.Body.DefaultHttps.EnableTLSv3
			}
		}
		modifyDefaultHttpsResp, err := d.sdkClient.ModifyDefaultHttpsWithContext(ctx, modifyDefaultHttpsReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'waf.ModifyDefaultHttps'", slog.Any("request", modifyDefaultHttpsReq), slog.Any("response", modifyDefaultHttpsResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ModifyDefaultHttps': %w", err)
		}
	} else {
		// 指定扩展域名，需替换扩展证书

		// 查询 CNAME 接入详情
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describedomaindetail
		describeDomainDetailReq := &aliwaf.DescribeDomainDetailRequest{
			RegionId:   tea.String(d.config.Region),
			InstanceId: tea.String(d.config.InstanceId),
			Domain:     tea.String(d.config.Domain),
		}
		describeDomainDetailResp, err := d.sdkClient.DescribeDomainDetailWithContext(ctx, describeDomainDetailReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'waf.DescribeDomainDetail'", slog.Any("request", describeDomainDetailReq), slog.Any("response", describeDomainDetailResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.DescribeDomainDetail': %w", err)
		}

		// 修改 CNAME 接入资源
		// REF: https://help.aliyun.com/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifydomain
		modifyDomainReq := &aliwaf.ModifyDomainRequest{
			RegionId:   tea.String(d.config.Region),
			InstanceId: tea.String(d.config.InstanceId),
			Domain:     tea.String(d.config.Domain),
			Listen:     &aliwaf.ModifyDomainRequestListen{CertId: tea.String(cloudCertId)},
			Redirect:   &aliwaf.ModifyDomainRequestRedirect{Loadbalance: tea.String("iphash")},
		}
		modifyDomainReq = _assign(modifyDomainReq, describeDomainDetailResp.Body)
		modifyDomainResp, err := d.sdkClient.ModifyDomainWithContext(ctx, modifyDomainReq, &dara.RuntimeOptions{})
		d.logger.Debug("sdk request 'waf.ModifyDomain'", slog.Any("request", modifyDomainReq), slog.Any("response", modifyDomainResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'waf.ModifyDomain': %w", err)
		}
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeySecret, region string) (*internal.WafClient, error) {
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

	client, err := internal.NewWafClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func _assign(source *aliwaf.ModifyDomainRequest, target *aliwaf.DescribeDomainDetailResponseBody) *aliwaf.ModifyDomainRequest {
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
			source.Listen.HttpPorts = lo.Map(target.Listen.HttpPorts, func(v *int64, _ int) *int32 {
				if v == nil {
					return nil
				}
				return tea.Int32(int32(*v))
			})
		}

		if target.Listen.HttpsPorts != nil {
			source.Listen.HttpsPorts = lo.Map(target.Listen.HttpsPorts, func(v *int64, _ int) *int32 {
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
			source.Redirect.Backends = lo.Map(target.Redirect.Backends, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectBackends, _ int) *string {
				if v == nil {
					return nil
				}
				return v.Backend
			})
		}

		if target.Redirect.BackupBackends != nil {
			source.Redirect.BackupBackends = lo.Map(target.Redirect.BackupBackends, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectBackupBackends, _ int) *string {
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
			source.Redirect.RequestHeaders = lo.Map(target.Redirect.RequestHeaders, func(v *aliwaf.DescribeDomainDetailResponseBodyRedirectRequestHeaders, _ int) *aliwaf.ModifyDomainRequestRedirectRequestHeaders {
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
