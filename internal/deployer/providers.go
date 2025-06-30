package deployer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	p1PanelConsole "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/1panel-console"
	p1PanelSite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/1panel-site"
	pAliyunALB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-alb"
	pAliyunAPIGW "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-apigw"
	pAliyunCAS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-cas"
	pAliyunCASDeploy "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-cas-deploy"
	pAliyunCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-cdn"
	pAliyunCLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-clb"
	pAliyunDCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-dcdn"
	pAliyunDDoS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-ddos"
	pAliyunESA "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-esa"
	pAliyunFC "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-fc"
	pAliyunGA "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-ga"
	pAliyunLive "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-live"
	pAliyunNLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-nlb"
	pAliyunOSS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-oss"
	pAliyunVOD "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-vod"
	pAliyunWAF "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aliyun-waf"
	pAPISIX "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/apisix"
	pAWSACM "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aws-acm"
	pAWSCloudFront "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aws-cloudfront"
	pAWSIAM "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/aws-iam"
	pAzureKeyVault "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/azure-keyvault"
	pBaiduCloudAppBLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baiducloud-appblb"
	pBaiduCloudBLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baiducloud-blb"
	pBaiduCloudCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baiducloud-cdn"
	pBaiduCloudCert "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baiducloud-cert"
	pBaishanCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baishan-cdn"
	pBaotaPanelConsole "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baotapanel-console"
	pBaotaPanelSite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baotapanel-site"
	pBaotaWAFConsole "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baotawaf-console"
	pBaotaWAFSite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/baotawaf-site"
	pBunnyCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/bunny-cdn"
	pBytePlusCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/byteplus-cdn"
	pCacheFly "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/cachefly"
	pCdnfly "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/cdnfly"
	pCTCCCloudAO "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-ao"
	pCTCCCloudCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-cdn"
	pCTCCCloudCMS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-cms"
	pCTCCCloudELB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-elb"
	pCTCCCloudICDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-icdn"
	pCTCCCloudLVDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ctcccloud-lvdn"
	pDogeCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/dogecloud-cdn"
	pEdgioApplications "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/edgio-applications"
	pFlexCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/flexcdn"
	pGcoreCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/gcore-cdn"
	pGoEdge "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/goedge"
	pHuaweiCloudCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/huaweicloud-cdn"
	pHuaweiCloudELB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/huaweicloud-elb"
	pHuaweiCloudSCM "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/huaweicloud-scm"
	pHuaweiCloudWAF "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/huaweicloud-waf"
	pJDCloudALB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/jdcloud-alb"
	pJDCloudCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/jdcloud-cdn"
	pJDCloudLive "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/jdcloud-live"
	pJDCloudVOD "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/jdcloud-vod"
	pK8sSecret "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/k8s-secret"
	pKong "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/kong"
	pLeCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/lecdn"
	pLocal "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/local"
	pNetlifySite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/netlify-site"
	pProxmoxVE "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/proxmoxve"
	pQiniuCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/qiniu-cdn"
	pQiniuPili "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/qiniu-pili"
	pRainYunRCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/rainyun-rcdn"
	pRatPanelConsole "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ratpanel-console"
	pRatPanelSite "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ratpanel-site"
	pSafeLine "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/safeline"
	pSSH "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ssh"
	pTencentCloudCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-cdn"
	pTencentCloudCLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-clb"
	pTencentCloudCOS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-cos"
	pTencentCloudCSS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-css"
	pTencentCloudECDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-ecdn"
	pTencentCloudEO "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-eo"
	pTencentCloudGAAP "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-gaap"
	pTencentCloudSCF "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-scf"
	pTencentCloudSSL "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-ssl"
	pTencentCloudSSLDeploy "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-ssl-deploy"
	pTencentCloudSSLUpdate "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-ssl-update"
	pTencentCloudVOD "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-vod"
	pTencentCloudWAF "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/tencentcloud-waf"
	pUCloudUCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ucloud-ucdn"
	pUCloudUS3 "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/ucloud-us3"
	pUniCloudWebHost "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/unicloud-webhost"
	pUpyunCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/upyun-cdn"
	pVolcEngineALB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-alb"
	pVolcEngineCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-cdn"
	pVolcEngineCertCenter "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-certcenter"
	pVolcEngineCLB "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-clb"
	pVolcEngineDCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-dcdn"
	pVolcEngineImageX "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-imagex"
	pVolcEngineLive "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-live"
	pVolcEngineTOS "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/volcengine-tos"
	pWangsuCDN "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/wangsu-cdn"
	pWangsuCDNPro "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/wangsu-cdnpro"
	pWangsuCertificate "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/wangsu-certificate"
	pWebhook "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/webhook"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
	xslices "github.com/certimate-go/certimate/pkg/utils/slices"
)

type deployerProviderOptions struct {
	Provider              domain.DeploymentProviderType
	ProviderAccessConfig  map[string]any
	ProviderServiceConfig map[string]any
}

func createSSLDeployerProvider(options *deployerProviderOptions) (core.SSLDeployer, error) {
	/*
	  注意：如果追加新的常量值，请保持以 ASCII 排序。
	  NOTICE: If you add new constant, please keep ASCII order.
	*/
	switch options.Provider {
	case domain.DeploymentProviderType1PanelConsole, domain.DeploymentProviderType1PanelSite:
		{
			access := domain.AccessConfigFor1Panel{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderType1PanelConsole:
				deployer, err := p1PanelConsole.NewSSLDeployerProvider(&p1PanelConsole.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiVersion:               access.ApiVersion,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
					AutoRestart:              xmaps.GetBool(options.ProviderServiceConfig, "autoRestart"),
				})
				return deployer, err

			case domain.DeploymentProviderType1PanelSite:
				deployer, err := p1PanelSite.NewSSLDeployerProvider(&p1PanelSite.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiVersion:               access.ApiVersion,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
					ResourceType:             p1PanelSite.ResourceType(xmaps.GetOrDefaultString(options.ProviderServiceConfig, "resourceType", string(p1PanelSite.RESOURCE_TYPE_WEBSITE))),
					WebsiteId:                xmaps.GetInt64(options.ProviderServiceConfig, "websiteId"),
					CertificateId:            xmaps.GetInt64(options.ProviderServiceConfig, "certificateId"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeAliyunALB, domain.DeploymentProviderTypeAliyunAPIGW, domain.DeploymentProviderTypeAliyunCAS, domain.DeploymentProviderTypeAliyunCASDeploy, domain.DeploymentProviderTypeAliyunCDN, domain.DeploymentProviderTypeAliyunCLB, domain.DeploymentProviderTypeAliyunDCDN, domain.DeploymentProviderTypeAliyunDDoS, domain.DeploymentProviderTypeAliyunESA, domain.DeploymentProviderTypeAliyunFC, domain.DeploymentProviderTypeAliyunGA, domain.DeploymentProviderTypeAliyunLive, domain.DeploymentProviderTypeAliyunNLB, domain.DeploymentProviderTypeAliyunOSS, domain.DeploymentProviderTypeAliyunVOD, domain.DeploymentProviderTypeAliyunWAF:
		{
			access := domain.AccessConfigForAliyun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeAliyunALB:
				deployer, err := pAliyunALB.NewSSLDeployerProvider(&pAliyunALB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pAliyunALB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunAPIGW:
				deployer, err := pAliyunAPIGW.NewSSLDeployerProvider(&pAliyunAPIGW.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ServiceType:     pAliyunAPIGW.ServiceType(xmaps.GetString(options.ProviderServiceConfig, "serviceType")),
					GatewayId:       xmaps.GetString(options.ProviderServiceConfig, "gatewayId"),
					GroupId:         xmaps.GetString(options.ProviderServiceConfig, "groupId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunCAS:
				deployer, err := pAliyunCAS.NewSSLDeployerProvider(&pAliyunCAS.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunCASDeploy:
				deployer, err := pAliyunCASDeploy.NewSSLDeployerProvider(&pAliyunCASDeploy.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceIds:     xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "resourceIds"), ";"), func(s string) bool { return s != "" }),
					ContactIds:      xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "contactIds"), ";"), func(s string) bool { return s != "" }),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunCDN:
				deployer, err := pAliyunCDN.NewSSLDeployerProvider(&pAliyunCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunCLB:
				deployer, err := pAliyunCLB.NewSSLDeployerProvider(&pAliyunCLB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pAliyunCLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerPort:    xmaps.GetOrDefaultInt32(options.ProviderServiceConfig, "listenerPort", 443),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunDCDN:
				deployer, err := pAliyunDCDN.NewSSLDeployerProvider(&pAliyunDCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunDDoS:
				deployer, err := pAliyunDDoS.NewSSLDeployerProvider(&pAliyunDDoS.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunESA:
				deployer, err := pAliyunESA.NewSSLDeployerProvider(&pAliyunESA.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					SiteId:          xmaps.GetInt64(options.ProviderServiceConfig, "siteId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunFC:
				deployer, err := pAliyunFC.NewSSLDeployerProvider(&pAliyunFC.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ServiceVersion:  xmaps.GetOrDefaultString(options.ProviderServiceConfig, "serviceVersion", "3.0"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunGA:
				deployer, err := pAliyunGA.NewSSLDeployerProvider(&pAliyunGA.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					ResourceType:    pAliyunGA.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					AcceleratorId:   xmaps.GetString(options.ProviderServiceConfig, "acceleratorId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunLive:
				deployer, err := pAliyunLive.NewSSLDeployerProvider(&pAliyunLive.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunNLB:
				deployer, err := pAliyunNLB.NewSSLDeployerProvider(&pAliyunNLB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pAliyunNLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunOSS:
				deployer, err := pAliyunOSS.NewSSLDeployerProvider(&pAliyunOSS.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					Bucket:          xmaps.GetString(options.ProviderServiceConfig, "bucket"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunVOD:
				deployer, err := pAliyunVOD.NewSSLDeployerProvider(&pAliyunVOD.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAliyunWAF:
				deployer, err := pAliyunWAF.NewSSLDeployerProvider(&pAliyunWAF.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ResourceGroupId: access.ResourceGroupId,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ServiceVersion:  xmaps.GetOrDefaultString(options.ProviderServiceConfig, "serviceVersion", "3.0"),
					InstanceId:      xmaps.GetString(options.ProviderServiceConfig, "instanceId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeAPISIX:
		{
			access := domain.AccessConfigForAPISIX{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pAPISIX.NewSSLDeployerProvider(&pAPISIX.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiKey:                   access.ApiKey,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pAPISIX.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				CertificateId:            xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeAWSACM, domain.DeploymentProviderTypeAWSCloudFront, domain.DeploymentProviderTypeAWSIAM:
		{
			access := domain.AccessConfigForAWS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeAWSACM:
				deployer, err := pAWSACM.NewSSLDeployerProvider(&pAWSACM.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					CertificateArn:  xmaps.GetString(options.ProviderServiceConfig, "certificateArn"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAWSCloudFront:
				deployer, err := pAWSCloudFront.NewSSLDeployerProvider(&pAWSCloudFront.SSLDeployerProviderConfig{
					AccessKeyId:       access.AccessKeyId,
					SecretAccessKey:   access.SecretAccessKey,
					Region:            xmaps.GetString(options.ProviderServiceConfig, "region"),
					DistributionId:    xmaps.GetString(options.ProviderServiceConfig, "distributionId"),
					CertificateSource: xmaps.GetOrDefaultString(options.ProviderServiceConfig, "certificateSource", "ACM"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeAWSIAM:
				deployer, err := pAWSIAM.NewSSLDeployerProvider(&pAWSIAM.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					CertificatePath: xmaps.GetOrDefaultString(options.ProviderServiceConfig, "certificatePath", "/"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeAzureKeyVault:
		{
			access := domain.AccessConfigForAzure{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeAzureKeyVault:
				deployer, err := pAzureKeyVault.NewSSLDeployerProvider(&pAzureKeyVault.SSLDeployerProviderConfig{
					TenantId:        access.TenantId,
					ClientId:        access.ClientId,
					ClientSecret:    access.ClientSecret,
					CloudName:       access.CloudName,
					KeyVaultName:    xmaps.GetString(options.ProviderServiceConfig, "keyvaultName"),
					CertificateName: xmaps.GetString(options.ProviderServiceConfig, "certificateName"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeBaiduCloudAppBLB, domain.DeploymentProviderTypeBaiduCloudBLB, domain.DeploymentProviderTypeBaiduCloudCDN, domain.DeploymentProviderTypeBaiduCloudCert:
		{
			access := domain.AccessConfigForBaiduCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeBaiduCloudAppBLB:
				deployer, err := pBaiduCloudAppBLB.NewSSLDeployerProvider(&pBaiduCloudAppBLB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pBaiduCloudAppBLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerPort:    xmaps.GetInt32(options.ProviderServiceConfig, "listenerPort"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeBaiduCloudBLB:
				deployer, err := pBaiduCloudBLB.NewSSLDeployerProvider(&pBaiduCloudBLB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pBaiduCloudBLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerPort:    xmaps.GetInt32(options.ProviderServiceConfig, "listenerPort"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeBaiduCloudCDN:
				deployer, err := pBaiduCloudCDN.NewSSLDeployerProvider(&pBaiduCloudCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeBaiduCloudCert:
				deployer, err := pBaiduCloudCert.NewSSLDeployerProvider(&pBaiduCloudCert.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeBaishanCDN:
		{
			access := domain.AccessConfigForBaishan{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeBaishanCDN:
				deployer, err := pBaishanCDN.NewSSLDeployerProvider(&pBaishanCDN.SSLDeployerProviderConfig{
					ApiToken:      access.ApiToken,
					Domain:        xmaps.GetString(options.ProviderServiceConfig, "domain"),
					CertificateId: xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeBaotaPanelConsole, domain.DeploymentProviderTypeBaotaPanelSite:
		{
			access := domain.AccessConfigForBaotaPanel{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeBaotaPanelConsole:
				deployer, err := pBaotaPanelConsole.NewSSLDeployerProvider(&pBaotaPanelConsole.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
					AutoRestart:              xmaps.GetBool(options.ProviderServiceConfig, "autoRestart"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeBaotaPanelSite:
				deployer, err := pBaotaPanelSite.NewSSLDeployerProvider(&pBaotaPanelSite.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
					SiteType:                 xmaps.GetOrDefaultString(options.ProviderServiceConfig, "siteType", "other"),
					SiteName:                 xmaps.GetString(options.ProviderServiceConfig, "siteName"),
					SiteNames:                xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "siteNames"), ";"), func(s string) bool { return s != "" }),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeBaotaWAFConsole, domain.DeploymentProviderTypeBaotaWAFSite:
		{
			access := domain.AccessConfigForBaotaWAF{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeBaotaWAFConsole:
				deployer, err := pBaotaWAFConsole.NewSSLDeployerProvider(&pBaotaWAFConsole.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
				})
				return deployer, err

			case domain.DeploymentProviderTypeBaotaWAFSite:
				deployer, err := pBaotaWAFSite.NewSSLDeployerProvider(&pBaotaWAFSite.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					ApiKey:                   access.ApiKey,
					AllowInsecureConnections: access.AllowInsecureConnections,
					SiteName:                 xmaps.GetString(options.ProviderServiceConfig, "siteName"),
					SitePort:                 xmaps.GetOrDefaultInt32(options.ProviderServiceConfig, "sitePort", 443),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeBunnyCDN:
		{
			access := domain.AccessConfigForBunny{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pBunnyCDN.NewSSLDeployerProvider(&pBunnyCDN.SSLDeployerProviderConfig{
				ApiKey:     access.ApiKey,
				PullZoneId: xmaps.GetString(options.ProviderServiceConfig, "pullZoneId"),
				Hostname:   xmaps.GetString(options.ProviderServiceConfig, "hostname"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeBytePlusCDN:
		{
			access := domain.AccessConfigForBytePlus{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeBytePlusCDN:
				deployer, err := pBytePlusCDN.NewSSLDeployerProvider(&pBytePlusCDN.SSLDeployerProviderConfig{
					AccessKey: access.AccessKey,
					SecretKey: access.SecretKey,
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeCacheFly:
		{
			access := domain.AccessConfigForCacheFly{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pCacheFly.NewSSLDeployerProvider(&pCacheFly.SSLDeployerProviderConfig{
				ApiToken: access.ApiToken,
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeCdnfly:
		{
			access := domain.AccessConfigForCdnfly{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pCdnfly.NewSSLDeployerProvider(&pCdnfly.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiKey:                   access.ApiKey,
				ApiSecret:                access.ApiSecret,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pCdnfly.ResourceType(xmaps.GetOrDefaultString(options.ProviderServiceConfig, "resourceType", string(pCdnfly.RESOURCE_TYPE_SITE))),
				SiteId:                   xmaps.GetString(options.ProviderServiceConfig, "siteId"),
				CertificateId:            xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeCTCCCloudAO, domain.DeploymentProviderTypeCTCCCloudCDN, domain.DeploymentProviderTypeCTCCCloudCMS, domain.DeploymentProviderTypeCTCCCloudELB, domain.DeploymentProviderTypeCTCCCloudICDN, domain.DeploymentProviderTypeCTCCCloudLVDN:
		{
			access := domain.AccessConfigForCTCCCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeCTCCCloudAO:
				deployer, err := pCTCCCloudAO.NewSSLDeployerProvider(&pCTCCCloudAO.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeCTCCCloudCDN:
				deployer, err := pCTCCCloudCDN.NewSSLDeployerProvider(&pCTCCCloudCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeCTCCCloudCMS:
				deployer, err := pCTCCCloudCMS.NewSSLDeployerProvider(&pCTCCCloudCMS.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
				})
				return deployer, err

			case domain.DeploymentProviderTypeCTCCCloudELB:
				deployer, err := pCTCCCloudELB.NewSSLDeployerProvider(&pCTCCCloudELB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					RegionId:        xmaps.GetString(options.ProviderServiceConfig, "regionId"),
					ResourceType:    pCTCCCloudELB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeCTCCCloudICDN:
				deployer, err := pCTCCCloudICDN.NewSSLDeployerProvider(&pCTCCCloudICDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeCTCCCloudLVDN:
				deployer, err := pCTCCCloudLVDN.NewSSLDeployerProvider(&pCTCCCloudLVDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					SecretAccessKey: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeDogeCloudCDN:
		{
			access := domain.AccessConfigForDogeCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pDogeCDN.NewSSLDeployerProvider(&pDogeCDN.SSLDeployerProviderConfig{
				AccessKey: access.AccessKey,
				SecretKey: access.SecretKey,
				Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeEdgioApplications:
		{
			access := domain.AccessConfigForEdgio{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pEdgioApplications.NewSSLDeployerProvider(&pEdgioApplications.SSLDeployerProviderConfig{
				ClientId:      access.ClientId,
				ClientSecret:  access.ClientSecret,
				EnvironmentId: xmaps.GetString(options.ProviderServiceConfig, "environmentId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeFlexCDN:
		{
			access := domain.AccessConfigForFlexCDN{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pFlexCDN.NewSSLDeployerProvider(&pFlexCDN.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiRole:                  access.ApiRole,
				AccessKeyId:              access.AccessKeyId,
				AccessKey:                access.AccessKey,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pFlexCDN.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				CertificateId:            xmaps.GetInt64(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeGcoreCDN:
		{
			access := domain.AccessConfigForGcore{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeGcoreCDN:
				deployer, err := pGcoreCDN.NewSSLDeployerProvider(&pGcoreCDN.SSLDeployerProviderConfig{
					ApiToken:      access.ApiToken,
					ResourceId:    xmaps.GetInt64(options.ProviderServiceConfig, "resourceId"),
					CertificateId: xmaps.GetInt64(options.ProviderServiceConfig, "certificateId"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeGoEdge:
		{
			access := domain.AccessConfigForGoEdge{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pGoEdge.NewSSLDeployerProvider(&pGoEdge.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiRole:                  access.ApiRole,
				AccessKeyId:              access.AccessKeyId,
				AccessKey:                access.AccessKey,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pGoEdge.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				CertificateId:            xmaps.GetInt64(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeHuaweiCloudCDN, domain.DeploymentProviderTypeHuaweiCloudELB, domain.DeploymentProviderTypeHuaweiCloudSCM, domain.DeploymentProviderTypeHuaweiCloudWAF:
		{
			access := domain.AccessConfigForHuaweiCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeHuaweiCloudCDN:
				deployer, err := pHuaweiCloudCDN.NewSSLDeployerProvider(&pHuaweiCloudCDN.SSLDeployerProviderConfig{
					AccessKeyId:         access.AccessKeyId,
					SecretAccessKey:     access.SecretAccessKey,
					EnterpriseProjectId: access.EnterpriseProjectId,
					Region:              xmaps.GetString(options.ProviderServiceConfig, "region"),
					Domain:              xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeHuaweiCloudELB:
				deployer, err := pHuaweiCloudELB.NewSSLDeployerProvider(&pHuaweiCloudELB.SSLDeployerProviderConfig{
					AccessKeyId:         access.AccessKeyId,
					SecretAccessKey:     access.SecretAccessKey,
					EnterpriseProjectId: access.EnterpriseProjectId,
					Region:              xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:        pHuaweiCloudELB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					CertificateId:       xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
					LoadbalancerId:      xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:          xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeHuaweiCloudSCM:
				deployer, err := pHuaweiCloudSCM.NewSSLDeployerProvider(&pHuaweiCloudSCM.SSLDeployerProviderConfig{
					AccessKeyId:         access.AccessKeyId,
					SecretAccessKey:     access.SecretAccessKey,
					EnterpriseProjectId: access.EnterpriseProjectId,
				})
				return deployer, err

			case domain.DeploymentProviderTypeHuaweiCloudWAF:
				deployer, err := pHuaweiCloudWAF.NewSSLDeployerProvider(&pHuaweiCloudWAF.SSLDeployerProviderConfig{
					AccessKeyId:         access.AccessKeyId,
					SecretAccessKey:     access.SecretAccessKey,
					EnterpriseProjectId: access.EnterpriseProjectId,
					Region:              xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:        pHuaweiCloudWAF.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					CertificateId:       xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
					Domain:              xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeJDCloudALB, domain.DeploymentProviderTypeJDCloudCDN, domain.DeploymentProviderTypeJDCloudLive, domain.DeploymentProviderTypeJDCloudVOD:
		{
			access := domain.AccessConfigForJDCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeJDCloudALB:
				deployer, err := pJDCloudALB.NewSSLDeployerProvider(&pJDCloudALB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					RegionId:        xmaps.GetString(options.ProviderServiceConfig, "regionId"),
					ResourceType:    pJDCloudALB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeJDCloudCDN:
				deployer, err := pJDCloudCDN.NewSSLDeployerProvider(&pJDCloudCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeJDCloudLive:
				deployer, err := pJDCloudLive.NewSSLDeployerProvider(&pJDCloudLive.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeJDCloudVOD:
				deployer, err := pJDCloudVOD.NewSSLDeployerProvider(&pJDCloudVOD.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeLeCDN:
		{
			access := domain.AccessConfigForLeCDN{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pLeCDN.NewSSLDeployerProvider(&pLeCDN.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiVersion:               access.ApiVersion,
				ApiRole:                  access.ApiRole,
				Username:                 access.Username,
				Password:                 access.Password,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pLeCDN.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				CertificateId:            xmaps.GetInt64(options.ProviderServiceConfig, "certificateId"),
				ClientId:                 xmaps.GetInt64(options.ProviderServiceConfig, "clientId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeLocal:
		{
			deployer, err := pLocal.NewSSLDeployerProvider(&pLocal.SSLDeployerProviderConfig{
				ShellEnv:                 pLocal.ShellEnvType(xmaps.GetString(options.ProviderServiceConfig, "shellEnv")),
				PreCommand:               xmaps.GetString(options.ProviderServiceConfig, "preCommand"),
				PostCommand:              xmaps.GetString(options.ProviderServiceConfig, "postCommand"),
				OutputFormat:             pLocal.OutputFormatType(xmaps.GetOrDefaultString(options.ProviderServiceConfig, "format", string(pLocal.OUTPUT_FORMAT_PEM))),
				OutputCertPath:           xmaps.GetString(options.ProviderServiceConfig, "certPath"),
				OutputServerCertPath:     xmaps.GetString(options.ProviderServiceConfig, "certPathForServerOnly"),
				OutputIntermediaCertPath: xmaps.GetString(options.ProviderServiceConfig, "certPathForIntermediaOnly"),
				OutputKeyPath:            xmaps.GetString(options.ProviderServiceConfig, "keyPath"),
				PfxPassword:              xmaps.GetString(options.ProviderServiceConfig, "pfxPassword"),
				JksAlias:                 xmaps.GetString(options.ProviderServiceConfig, "jksAlias"),
				JksKeypass:               xmaps.GetString(options.ProviderServiceConfig, "jksKeypass"),
				JksStorepass:             xmaps.GetString(options.ProviderServiceConfig, "jksStorepass"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeKong:
		{
			access := domain.AccessConfigForKong{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pKong.NewSSLDeployerProvider(&pKong.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiToken:                 access.ApiToken,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pKong.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				Workspace:                xmaps.GetString(options.ProviderServiceConfig, "workspace"),
				CertificateId:            xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeKubernetesSecret:
		{
			access := domain.AccessConfigForKubernetes{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pK8sSecret.NewSSLDeployerProvider(&pK8sSecret.SSLDeployerProviderConfig{
				KubeConfig:          access.KubeConfig,
				Namespace:           xmaps.GetOrDefaultString(options.ProviderServiceConfig, "namespace", "default"),
				SecretName:          xmaps.GetString(options.ProviderServiceConfig, "secretName"),
				SecretType:          xmaps.GetOrDefaultString(options.ProviderServiceConfig, "secretType", "kubernetes.io/tls"),
				SecretDataKeyForCrt: xmaps.GetOrDefaultString(options.ProviderServiceConfig, "secretDataKeyForCrt", "tls.crt"),
				SecretDataKeyForKey: xmaps.GetOrDefaultString(options.ProviderServiceConfig, "secretDataKeyForKey", "tls.key"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeNetlifySite:
		{
			access := domain.AccessConfigForNetlify{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pNetlifySite.NewSSLDeployerProvider(&pNetlifySite.SSLDeployerProviderConfig{
				ApiToken: access.ApiToken,
				SiteId:   xmaps.GetString(options.ProviderServiceConfig, "siteId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeProxmoxVE:
		{
			access := domain.AccessConfigForProxmoxVE{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pProxmoxVE.NewSSLDeployerProvider(&pProxmoxVE.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiToken:                 access.ApiToken,
				ApiTokenSecret:           access.ApiTokenSecret,
				AllowInsecureConnections: access.AllowInsecureConnections,
				NodeName:                 xmaps.GetString(options.ProviderServiceConfig, "nodeName"),
				AutoRestart:              xmaps.GetBool(options.ProviderServiceConfig, "autoRestart"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeQiniuCDN, domain.DeploymentProviderTypeQiniuKodo, domain.DeploymentProviderTypeQiniuPili:
		{
			access := domain.AccessConfigForQiniu{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeQiniuCDN, domain.DeploymentProviderTypeQiniuKodo:
				deployer, err := pQiniuCDN.NewSSLDeployerProvider(&pQiniuCDN.SSLDeployerProviderConfig{
					AccessKey: access.AccessKey,
					SecretKey: access.SecretKey,
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeQiniuPili:
				deployer, err := pQiniuPili.NewSSLDeployerProvider(&pQiniuPili.SSLDeployerProviderConfig{
					AccessKey: access.AccessKey,
					SecretKey: access.SecretKey,
					Hub:       xmaps.GetString(options.ProviderServiceConfig, "hub"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeRainYunRCDN:
		{
			access := domain.AccessConfigForRainYun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeRainYunRCDN:
				deployer, err := pRainYunRCDN.NewSSLDeployerProvider(&pRainYunRCDN.SSLDeployerProviderConfig{
					ApiKey:     access.ApiKey,
					InstanceId: xmaps.GetInt32(options.ProviderServiceConfig, "instanceId"),
					Domain:     xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeRatPanelConsole, domain.DeploymentProviderTypeRatPanelSite:
		{
			access := domain.AccessConfigForRatPanel{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeRatPanelConsole:
				deployer, err := pRatPanelConsole.NewSSLDeployerProvider(&pRatPanelConsole.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					AccessTokenId:            access.AccessTokenId,
					AccessToken:              access.AccessToken,
					AllowInsecureConnections: access.AllowInsecureConnections,
				})
				return deployer, err

			case domain.DeploymentProviderTypeRatPanelSite:
				deployer, err := pRatPanelSite.NewSSLDeployerProvider(&pRatPanelSite.SSLDeployerProviderConfig{
					ServerUrl:                access.ServerUrl,
					AccessTokenId:            access.AccessTokenId,
					AccessToken:              access.AccessToken,
					AllowInsecureConnections: access.AllowInsecureConnections,
					SiteName:                 xmaps.GetString(options.ProviderServiceConfig, "siteName"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeSafeLine:
		{
			access := domain.AccessConfigForSafeLine{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pSafeLine.NewSSLDeployerProvider(&pSafeLine.SSLDeployerProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiToken:                 access.ApiToken,
				AllowInsecureConnections: access.AllowInsecureConnections,
				ResourceType:             pSafeLine.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
				CertificateId:            xmaps.GetInt32(options.ProviderServiceConfig, "certificateId"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeSSH:
		{
			access := domain.AccessConfigForSSH{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			jumpServers := make([]pSSH.JumpServerConfig, len(access.JumpServers))
			for i, jumpServer := range access.JumpServers {
				jumpServers[i] = pSSH.JumpServerConfig{
					SshHost:          jumpServer.Host,
					SshPort:          jumpServer.Port,
					SshAuthMethod:    jumpServer.AuthMethod,
					SshUsername:      jumpServer.Username,
					SshPassword:      jumpServer.Password,
					SshKey:           jumpServer.Key,
					SshKeyPassphrase: jumpServer.KeyPassphrase,
				}
			}

			deployer, err := pSSH.NewSSLDeployerProvider(&pSSH.SSLDeployerProviderConfig{
				SshHost:                  access.Host,
				SshPort:                  access.Port,
				SshAuthMethod:            access.AuthMethod,
				SshUsername:              access.Username,
				SshPassword:              access.Password,
				SshKey:                   access.Key,
				SshKeyPassphrase:         access.KeyPassphrase,
				JumpServers:              jumpServers,
				UseSCP:                   xmaps.GetBool(options.ProviderServiceConfig, "useSCP"),
				PreCommand:               xmaps.GetString(options.ProviderServiceConfig, "preCommand"),
				PostCommand:              xmaps.GetString(options.ProviderServiceConfig, "postCommand"),
				OutputFormat:             pSSH.OutputFormatType(xmaps.GetOrDefaultString(options.ProviderServiceConfig, "format", string(pSSH.OUTPUT_FORMAT_PEM))),
				OutputCertPath:           xmaps.GetString(options.ProviderServiceConfig, "certPath"),
				OutputServerCertPath:     xmaps.GetString(options.ProviderServiceConfig, "certPathForServerOnly"),
				OutputIntermediaCertPath: xmaps.GetString(options.ProviderServiceConfig, "certPathForIntermediaOnly"),
				OutputKeyPath:            xmaps.GetString(options.ProviderServiceConfig, "keyPath"),
				PfxPassword:              xmaps.GetString(options.ProviderServiceConfig, "pfxPassword"),
				JksAlias:                 xmaps.GetString(options.ProviderServiceConfig, "jksAlias"),
				JksKeypass:               xmaps.GetString(options.ProviderServiceConfig, "jksKeypass"),
				JksStorepass:             xmaps.GetString(options.ProviderServiceConfig, "jksStorepass"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeTencentCloudCDN, domain.DeploymentProviderTypeTencentCloudCLB, domain.DeploymentProviderTypeTencentCloudCOS, domain.DeploymentProviderTypeTencentCloudCSS, domain.DeploymentProviderTypeTencentCloudECDN, domain.DeploymentProviderTypeTencentCloudEO, domain.DeploymentProviderTypeTencentCloudGAAP, domain.DeploymentProviderTypeTencentCloudSCF, domain.DeploymentProviderTypeTencentCloudSSL, domain.DeploymentProviderTypeTencentCloudSSLDeploy, domain.DeploymentProviderTypeTencentCloudSSLUpdate, domain.DeploymentProviderTypeTencentCloudVOD, domain.DeploymentProviderTypeTencentCloudWAF:
		{
			access := domain.AccessConfigForTencentCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeTencentCloudCDN:
				deployer, err := pTencentCloudCDN.NewSSLDeployerProvider(&pTencentCloudCDN.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudCLB:
				deployer, err := pTencentCloudCLB.NewSSLDeployerProvider(&pTencentCloudCLB.SSLDeployerProviderConfig{
					SecretId:       access.SecretId,
					SecretKey:      access.SecretKey,
					Endpoint:       xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Region:         xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:   pTencentCloudCLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId: xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:     xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
					Domain:         xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudCOS:
				deployer, err := pTencentCloudCOS.NewSSLDeployerProvider(&pTencentCloudCOS.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Region:    xmaps.GetString(options.ProviderServiceConfig, "region"),
					Bucket:    xmaps.GetString(options.ProviderServiceConfig, "bucket"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudCSS:
				deployer, err := pTencentCloudCSS.NewSSLDeployerProvider(&pTencentCloudCSS.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudECDN:
				deployer, err := pTencentCloudECDN.NewSSLDeployerProvider(&pTencentCloudECDN.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudEO:
				deployer, err := pTencentCloudEO.NewSSLDeployerProvider(&pTencentCloudEO.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					ZoneId:    xmaps.GetString(options.ProviderServiceConfig, "zoneId"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudGAAP:
				deployer, err := pTencentCloudGAAP.NewSSLDeployerProvider(&pTencentCloudGAAP.SSLDeployerProviderConfig{
					SecretId:     access.SecretId,
					SecretKey:    access.SecretKey,
					Endpoint:     xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					ResourceType: pTencentCloudGAAP.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					ProxyId:      xmaps.GetString(options.ProviderServiceConfig, "proxyId"),
					ListenerId:   xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudSCF:
				deployer, err := pTencentCloudSCF.NewSSLDeployerProvider(&pTencentCloudSCF.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Region:    xmaps.GetString(options.ProviderServiceConfig, "region"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudSSL:
				deployer, err := pTencentCloudSSL.NewSSLDeployerProvider(&pTencentCloudSSL.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudSSLDeploy:
				deployer, err := pTencentCloudSSLDeploy.NewSSLDeployerProvider(&pTencentCloudSSLDeploy.SSLDeployerProviderConfig{
					SecretId:     access.SecretId,
					SecretKey:    access.SecretKey,
					Endpoint:     xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Region:       xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType: xmaps.GetString(options.ProviderServiceConfig, "resourceType"),
					ResourceIds:  xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "resourceIds"), ";"), func(s string) bool { return s != "" }),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudSSLUpdate:
				deployer, err := pTencentCloudSSLUpdate.NewSSLDeployerProvider(&pTencentCloudSSLUpdate.SSLDeployerProviderConfig{
					SecretId:        access.SecretId,
					SecretKey:       access.SecretKey,
					Endpoint:        xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					CertificateId:   xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
					IsReplaced:      xmaps.GetBool(options.ProviderServiceConfig, "isReplaced"),
					ResourceTypes:   xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "resourceTypes"), ";"), func(s string) bool { return s != "" }),
					ResourceRegions: xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "resourceRegions"), ";"), func(s string) bool { return s != "" }),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudVOD:
				deployer, err := pTencentCloudVOD.NewSSLDeployerProvider(&pTencentCloudVOD.SSLDeployerProviderConfig{
					SecretId:  access.SecretId,
					SecretKey: access.SecretKey,
					Endpoint:  xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					SubAppId:  xmaps.GetInt64(options.ProviderServiceConfig, "subAppId"),
					Domain:    xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeTencentCloudWAF:
				deployer, err := pTencentCloudWAF.NewSSLDeployerProvider(&pTencentCloudWAF.SSLDeployerProviderConfig{
					SecretId:   access.SecretId,
					SecretKey:  access.SecretKey,
					Endpoint:   xmaps.GetString(options.ProviderServiceConfig, "endpoint"),
					Domain:     xmaps.GetString(options.ProviderServiceConfig, "domain"),
					DomainId:   xmaps.GetString(options.ProviderServiceConfig, "domainId"),
					InstanceId: xmaps.GetString(options.ProviderServiceConfig, "instanceId"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeUCloudUCDN, domain.DeploymentProviderTypeUCloudUS3:
		{
			access := domain.AccessConfigForUCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeUCloudUCDN:
				deployer, err := pUCloudUCDN.NewSSLDeployerProvider(&pUCloudUCDN.SSLDeployerProviderConfig{
					PrivateKey: access.PrivateKey,
					PublicKey:  access.PublicKey,
					ProjectId:  access.ProjectId,
					DomainId:   xmaps.GetString(options.ProviderServiceConfig, "domainId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeUCloudUS3:
				deployer, err := pUCloudUS3.NewSSLDeployerProvider(&pUCloudUS3.SSLDeployerProviderConfig{
					PrivateKey: access.PrivateKey,
					PublicKey:  access.PublicKey,
					ProjectId:  access.ProjectId,
					Region:     xmaps.GetString(options.ProviderServiceConfig, "region"),
					Bucket:     xmaps.GetString(options.ProviderServiceConfig, "bucket"),
					Domain:     xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeUniCloudWebHost:
		{
			access := domain.AccessConfigForUniCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			deployer, err := pUniCloudWebHost.NewSSLDeployerProvider(&pUniCloudWebHost.SSLDeployerProviderConfig{
				Username:      access.Username,
				Password:      access.Password,
				SpaceProvider: xmaps.GetString(options.ProviderServiceConfig, "spaceProvider"),
				SpaceId:       xmaps.GetString(options.ProviderServiceConfig, "spaceId"),
				Domain:        xmaps.GetString(options.ProviderServiceConfig, "domain"),
			})
			return deployer, err
		}

	case domain.DeploymentProviderTypeUpyunCDN, domain.DeploymentProviderTypeUpyunFile:
		{
			access := domain.AccessConfigForUpyun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeUpyunCDN, domain.DeploymentProviderTypeUpyunFile:
				deployer, err := pUpyunCDN.NewSSLDeployerProvider(&pUpyunCDN.SSLDeployerProviderConfig{
					Username: access.Username,
					Password: access.Password,
					Domain:   xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeVolcEngineALB, domain.DeploymentProviderTypeVolcEngineCDN, domain.DeploymentProviderTypeVolcEngineCertCenter, domain.DeploymentProviderTypeVolcEngineCLB, domain.DeploymentProviderTypeVolcEngineDCDN, domain.DeploymentProviderTypeVolcEngineImageX, domain.DeploymentProviderTypeVolcEngineLive, domain.DeploymentProviderTypeVolcEngineTOS:
		{
			access := domain.AccessConfigForVolcEngine{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeVolcEngineALB:
				deployer, err := pVolcEngineALB.NewSSLDeployerProvider(&pVolcEngineALB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pVolcEngineALB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineCDN:
				deployer, err := pVolcEngineCDN.NewSSLDeployerProvider(&pVolcEngineCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineCertCenter:
				deployer, err := pVolcEngineCertCenter.NewSSLDeployerProvider(&pVolcEngineCertCenter.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineCLB:
				deployer, err := pVolcEngineCLB.NewSSLDeployerProvider(&pVolcEngineCLB.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ResourceType:    pVolcEngineCLB.ResourceType(xmaps.GetString(options.ProviderServiceConfig, "resourceType")),
					LoadbalancerId:  xmaps.GetString(options.ProviderServiceConfig, "loadbalancerId"),
					ListenerId:      xmaps.GetString(options.ProviderServiceConfig, "listenerId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineDCDN:
				deployer, err := pVolcEngineDCDN.NewSSLDeployerProvider(&pVolcEngineDCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineImageX:
				deployer, err := pVolcEngineImageX.NewSSLDeployerProvider(&pVolcEngineImageX.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					ServiceId:       xmaps.GetString(options.ProviderServiceConfig, "serviceId"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineLive:
				deployer, err := pVolcEngineLive.NewSSLDeployerProvider(&pVolcEngineLive.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeVolcEngineTOS:
				deployer, err := pVolcEngineTOS.NewSSLDeployerProvider(&pVolcEngineTOS.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.SecretAccessKey,
					Region:          xmaps.GetString(options.ProviderServiceConfig, "region"),
					Bucket:          xmaps.GetString(options.ProviderServiceConfig, "bucket"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeWangsuCDN, domain.DeploymentProviderTypeWangsuCDNPro, domain.DeploymentProviderTypeWangsuCertificate:
		{
			access := domain.AccessConfigForWangsu{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.DeploymentProviderTypeWangsuCDN:
				deployer, err := pWangsuCDN.NewSSLDeployerProvider(&pWangsuCDN.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					Domains:         xslices.Filter(strings.Split(xmaps.GetString(options.ProviderServiceConfig, "domains"), ";"), func(s string) bool { return s != "" }),
				})
				return deployer, err

			case domain.DeploymentProviderTypeWangsuCDNPro:
				deployer, err := pWangsuCDNPro.NewSSLDeployerProvider(&pWangsuCDNPro.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					ApiKey:          access.ApiKey,
					Environment:     xmaps.GetOrDefaultString(options.ProviderServiceConfig, "environment", "production"),
					Domain:          xmaps.GetString(options.ProviderServiceConfig, "domain"),
					CertificateId:   xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
					WebhookId:       xmaps.GetString(options.ProviderServiceConfig, "webhookId"),
				})
				return deployer, err

			case domain.DeploymentProviderTypeWangsuCertificate:
				deployer, err := pWangsuCertificate.NewSSLDeployerProvider(&pWangsuCertificate.SSLDeployerProviderConfig{
					AccessKeyId:     access.AccessKeyId,
					AccessKeySecret: access.AccessKeySecret,
					CertificateId:   xmaps.GetString(options.ProviderServiceConfig, "certificateId"),
				})
				return deployer, err

			default:
				break
			}
		}

	case domain.DeploymentProviderTypeWebhook:
		{
			access := domain.AccessConfigForWebhook{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			mergedHeaders := make(map[string]string)
			if defaultHeadersString := access.HeadersString; defaultHeadersString != "" {
				h, err := xhttp.ParseHeaders(defaultHeadersString)
				if err != nil {
					return nil, fmt.Errorf("failed to parse webhook headers: %w", err)
				}
				for key := range h {
					mergedHeaders[http.CanonicalHeaderKey(key)] = h.Get(key)
				}
			}
			if extendedHeadersString := xmaps.GetString(options.ProviderServiceConfig, "headers"); extendedHeadersString != "" {
				h, err := xhttp.ParseHeaders(extendedHeadersString)
				if err != nil {
					return nil, fmt.Errorf("failed to parse webhook headers: %w", err)
				}
				for key := range h {
					mergedHeaders[http.CanonicalHeaderKey(key)] = h.Get(key)
				}
			}

			deployer, err := pWebhook.NewSSLDeployerProvider(&pWebhook.SSLDeployerProviderConfig{
				WebhookUrl:               access.Url,
				WebhookData:              xmaps.GetOrDefaultString(options.ProviderServiceConfig, "webhookData", access.DefaultDataForDeployment),
				Method:                   access.Method,
				Headers:                  mergedHeaders,
				AllowInsecureConnections: access.AllowInsecureConnections,
			})
			return deployer, err
		}
	}

	return nil, fmt.Errorf("unsupported deployer provider '%s'", string(options.Provider))
}
