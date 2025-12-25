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

	certId, ok := upres.ExtendedData["CertIdentifier"].(string)
	if !ok || certId == "" {
		return fmt.Errorf("invalid certificate identifier in upload response")
	}
	// 根据接入方式决定部署方式
	switch d.config.ServiceType {
	case SERVICE_TYPE_CLOUDRESOURCE:
		if err := d.deployToWAF3WithCloudResource(ctx, certId); err != nil {
			return err
		}

	case SERVICE_TYPE_CNAME:
		if err := d.deployToWAF3WithCNAME(ctx, certId); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported service version '%s'", d.config.ServiceVersion)
	}

	return nil
}

func (d *Deployer) deployToWAF3WithCloudResource(ctx context.Context, cloudCertId string) error {
	// 参数校验
	if d.config.ResourceProduct == "" {
		return errors.New("config `resourceProduct` is required")
	}
	if d.config.ResourceId == "" {
		return errors.New("config `resourceId` is required")
	}
	// 默认端口为 443
	if d.config.ResourcePort == 0 {
		d.config.ResourcePort = 443
	}

	// DescribeProductInstances
	// 查询云产品实例列表，包含实例的基本信息和接入端口等信息。
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describeproductinstances
	var resourceInstance *aliwaf.DescribeProductInstancesResponseBodyProductInstances
	var resourceInstancePort *aliwaf.DescribeProductInstancesResponseBodyProductInstancesAccessPortAndProtocols
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

		resourceInstancePort, _ = lo.Find(resourceInstance.AccessPortAndProtocols, func(p *aliwaf.DescribeProductInstancesResponseBodyProductInstancesAccessPortAndProtocols) bool {
			return tea.Int32Value(p.Port) == d.config.ResourcePort
		})
		if resourceInstancePort == nil {
			return fmt.Errorf("could not find waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
		}
	}

	// DescribeResourceInstanceCerts
	// 查询云产品实例已同步的证书列表。
	// 这个接口返回的证书列表里包含了所有已同步的证书信息，包括证书 ID、证书公共名称、过期时间等信息。
	// 但是不会区分哪些是默认证书，哪些是扩展证书。
	// 而且此接口可能会返回当前 WAF 下所有的证书，而不只是指定的的云实例使用的证书。
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describeresourceinstancecerts
	var resourceInstanceCertificates = make([]*aliwaf.DescribeResourceInstanceCertsResponseBodyCerts, 0)

	describeResourceInstanceCertsReq := &aliwaf.DescribeResourceInstanceCertsRequest{
		RegionId:                       tea.String(d.config.Region),
		ResourceManagerResourceGroupId: lo.EmptyableToPtr(d.config.ResourceGroupId),
		InstanceId:                     tea.String(d.config.InstanceId),
		ResourceInstanceId:             tea.String(d.config.ResourceId),
	}
	describeResourceInstanceCertsResp, err := d.sdkClient.DescribeResourceInstanceCertsWithContext(ctx, describeResourceInstanceCertsReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'waf.DescribeResourceInstanceCerts'", slog.Any("request", describeResourceInstanceCertsReq), slog.Any("response", describeResourceInstanceCertsResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.DescribeResourceInstanceCerts': %w", err)
	}

	if describeResourceInstanceCertsResp.Body == nil {
		return fmt.Errorf("describe waf %s cloud resource instance certificates response body is empty", d.config.InstanceId)
	}

	resourceInstanceCertificates = describeResourceInstanceCertsResp.Body.Certs
	if len(resourceInstanceCertificates) == 0 {
		return fmt.Errorf("describe waf %s cloud resource instance certificates response certs list is empty", d.config.InstanceId)
	}

	// DescribeCloudResourceAccessPortDetails
	// 获取云产品实例的接入端口详情，包含证书列表、实例 ID 等信息。
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-describecloudresourceaccessportdetails
	var cloudResourceAccessPortDetailsCertificates = make([]*aliwaf.DescribeCloudResourceAccessPortDetailsResponseBodyAccessPortDetailsCertificates, 0)
	describeCloudResourceAccessPortDetailsRequest := &aliwaf.DescribeCloudResourceAccessPortDetailsRequest{
		RegionId:           tea.String(d.config.Region),
		InstanceId:         tea.String(d.config.InstanceId),
		ResourceInstanceId: tea.String(d.config.ResourceId),
		Port:               tea.String(fmt.Sprintf("%d", d.config.ResourcePort)),
	}
	describeCloudResourceAccessPortDetailsResponse, err := d.sdkClient.DescribeCloudResourceAccessPortDetailsWithContext(ctx, describeCloudResourceAccessPortDetailsRequest, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'waf.DescribeCloudResourceAccessPortDetails'", slog.Any("request", describeCloudResourceAccessPortDetailsRequest), slog.Any("response", describeCloudResourceAccessPortDetailsResponse))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.DescribeCloudResourceAccessPortDetails': %w", err)
	}

	if describeCloudResourceAccessPortDetailsResponse.Body.AccessPortDetails == nil {
		return fmt.Errorf("could not get access port details of waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
	}

	cloudResourceAccessPortDetailsCertificates = describeCloudResourceAccessPortDetailsResponse.Body.AccessPortDetails[0].Certificates
	// 至少应该存在一张默认证书
	if len(cloudResourceAccessPortDetailsCertificates) == 0 {
		return fmt.Errorf("could not get certificates of waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
	}

	// 获取由 WAF 自动生成的 CloudResourceId，后续更新证书配置时需要用到
	cloudResourceId := describeCloudResourceAccessPortDetailsResponse.Body.AccessPortDetails[0].CloudResourceId
	if cloudResourceId == nil || tea.StringValue(cloudResourceId) == "" {
		return fmt.Errorf("could not get cloud resource id of waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
	}

	// ModifyCloudResourceCert
	// 更新证书配置
	// REF: https://www.alibabacloud.com/help/zh/waf/web-application-firewall-3-0/developer-reference/api-waf-openapi-2021-10-01-modifycloudresourcecert
	modifyCloudResourceCertReq := &aliwaf.ModifyCloudResourceCertRequest{
		RegionId:        tea.String(d.config.Region),
		InstanceId:      tea.String(d.config.InstanceId),
		CloudResourceId: cloudResourceId,
	}

	if d.config.Domain == "" {
		// 没有指定扩展域名，只需替换默认证书
		const certAppliedTypeDefault = "default"

		// 从已有列表里找出默认证书 ID，替换为新的证书 ID，其他的扩展证书保持不变（如果有）
		modifyCloudResourceCertReq.Certificates = lo.Map(cloudResourceAccessPortDetailsCertificates, func(c *aliwaf.DescribeCloudResourceAccessPortDetailsResponseBodyAccessPortDetailsCertificates, _ int) *aliwaf.ModifyCloudResourceCertRequestCertificates {
			appliedType := tea.StringValue(c.AppliedType)
			certId := tea.StringValue(c.CertificateId)
			if appliedType == certAppliedTypeDefault {
				certId = cloudCertId
			}
			return &aliwaf.ModifyCloudResourceCertRequestCertificates{
				CertificateId: tea.String(certId),
				AppliedType:   tea.String(appliedType),
			}
		})
	} else {
		// 指定了扩展域名，如果已有就替换，没有就新增扩展证书
		const certAppliedTypeExtension = "extension"

		// 先拿到当前实例已有的扩展证书的列表
		// 构建新的证书列表，保留默认证书，处理扩展证书的替换或保留
		var found bool
		newCertificates := make([]*aliwaf.ModifyCloudResourceCertRequestCertificates, 0, len(cloudResourceAccessPortDetailsCertificates)+1)
		for _, c := range cloudResourceAccessPortDetailsCertificates {
			applied := tea.StringValue(c.AppliedType)
			certId := tea.StringValue(c.CertificateId)

			if applied == certAppliedTypeExtension {
				// 从已同步的云产品实例全部证书列表里找到对应的证书，获取 CommonName
				var certCommonName string
				for _, rc := range resourceInstanceCertificates {
					if tea.StringValue(rc.CertIdentifier) == certId {
						certCommonName = tea.StringValue(rc.CommonName)
						break
					}
				}
				if certCommonName == d.config.Domain {
					// 找到和目标域名匹配的扩展证书，替换为新证书 ID
					newCertificates = append(newCertificates, &aliwaf.ModifyCloudResourceCertRequestCertificates{
						CertificateId: tea.String(cloudCertId),
						AppliedType:   tea.String(certAppliedTypeExtension),
					})
					found = true
				} else {
					// 保留原来的扩展证书
					newCertificates = append(newCertificates, &aliwaf.ModifyCloudResourceCertRequestCertificates{
						CertificateId: c.CertificateId,
						AppliedType:   c.AppliedType,
					})
				}
			} else {
				// 默认证书保持不变
				newCertificates = append(newCertificates, &aliwaf.ModifyCloudResourceCertRequestCertificates{
					CertificateId: c.CertificateId,
					AppliedType:   c.AppliedType,
				})
			}
		}
		if !found {
			// 如果没有找到要替换的扩展证书，追加一个扩展证书
			newCertificates = append(newCertificates, &aliwaf.ModifyCloudResourceCertRequestCertificates{
				CertificateId: tea.String(cloudCertId),
				AppliedType:   tea.String(certAppliedTypeExtension),
			})
		}

		modifyCloudResourceCertReq.Certificates = newCertificates
	}

	// 过滤掉不存在或已过期的证书，防止接口报错
	modifyCloudResourceCertReq.Certificates = lo.Filter(modifyCloudResourceCertReq.Certificates, func(c *aliwaf.ModifyCloudResourceCertRequestCertificates, _ int) bool {
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

	// 最后检查待更新的列表里是否存在默认证书，没有则退出，否则接口会报错
	var hasDefaultCert bool
	for _, c := range modifyCloudResourceCertReq.Certificates {
		if tea.StringValue(c.AppliedType) == "default" {
			hasDefaultCert = true
			break
		}
	}
	if !hasDefaultCert {
		return fmt.Errorf("could not find valid default certificate to apply for waf '%s' cloud resource '%s %s:%d'", d.config.InstanceId, d.config.ResourceProduct, d.config.ResourceId, d.config.ResourcePort)
	}

	// 执行更新请求
	modifyCloudResourceCertResp, err := d.sdkClient.ModifyCloudResourceCertWithContext(ctx, modifyCloudResourceCertReq, &dara.RuntimeOptions{})
	d.logger.Debug("sdk request 'waf.ModifyCloudResourceCert'", slog.Any("request", modifyCloudResourceCertReq), slog.Any("response", modifyCloudResourceCertResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'waf.ModifyCloudResourceCert': %w", err)
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
