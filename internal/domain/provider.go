package domain

type AccessProviderType string

/*
授权提供商类型常量值。

	注意：如果追加新的常量值，请保持以 ASCII 排序。
	NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	AccessProviderType1Panel              = AccessProviderType("1panel")
	AccessProviderTypeACMECA              = AccessProviderType("acmeca")
	AccessProviderTypeACMEHttpReq         = AccessProviderType("acmehttpreq")
	AccessProviderTypeAkamai              = AccessProviderType("akamai") // Akamai（预留）
	AccessProviderTypeAliyun              = AccessProviderType("aliyun")
	AccessProviderTypeAPISIX              = AccessProviderType("apisix")
	AccessProviderTypeAWS                 = AccessProviderType("aws")
	AccessProviderTypeAzure               = AccessProviderType("azure")
	AccessProviderTypeBaiduCloud          = AccessProviderType("baiducloud")
	AccessProviderTypeBaishan             = AccessProviderType("baishan")
	AccessProviderTypeBaotaPanel          = AccessProviderType("baotapanel")
	AccessProviderTypeBaotaWAF            = AccessProviderType("baotawaf")
	AccessProviderTypeBytePlus            = AccessProviderType("byteplus")
	AccessProviderTypeBunny               = AccessProviderType("bunny")
	AccessProviderTypeBuypass             = AccessProviderType("buypass")
	AccessProviderTypeCacheFly            = AccessProviderType("cachefly")
	AccessProviderTypeCdnfly              = AccessProviderType("cdnfly")
	AccessProviderTypeCloudflare          = AccessProviderType("cloudflare")
	AccessProviderTypeClouDNS             = AccessProviderType("cloudns")
	AccessProviderTypeCMCCCloud           = AccessProviderType("cmcccloud")
	AccessProviderTypeConstellix          = AccessProviderType("constellix")
	AccessProviderTypeCTCCCloud           = AccessProviderType("ctcccloud")
	AccessProviderTypeCUCCCloud           = AccessProviderType("cucccloud") // 联通云（预留）
	AccessProviderTypeDeSEC               = AccessProviderType("desec")
	AccessProviderTypeDigitalOcean        = AccessProviderType("digitalocean")
	AccessProviderTypeDingTalkBot         = AccessProviderType("dingtalkbot")
	AccessProviderTypeDiscordBot          = AccessProviderType("discordbot")
	AccessProviderTypeDNSLA               = AccessProviderType("dnsla")
	AccessProviderTypeDogeCloud           = AccessProviderType("dogecloud")
	AccessProviderTypeDuckDNS             = AccessProviderType("duckdns")
	AccessProviderTypeDynv6               = AccessProviderType("dynv6")
	AccessProviderTypeEdgio               = AccessProviderType("edgio")
	AccessProviderTypeEmail               = AccessProviderType("email")
	AccessProviderTypeFastly              = AccessProviderType("fastly") // Fastly（预留）
	AccessProviderTypeFlexCDN             = AccessProviderType("flexcdn")
	AccessProviderTypeGname               = AccessProviderType("gname")
	AccessProviderTypeGcore               = AccessProviderType("gcore")
	AccessProviderTypeGoDaddy             = AccessProviderType("godaddy")
	AccessProviderTypeGoEdge              = AccessProviderType("goedge")
	AccessProviderTypeGoogleTrustServices = AccessProviderType("googletrustservices")
	AccessProviderTypeHetzner             = AccessProviderType("hetzner")
	AccessProviderTypeHuaweiCloud         = AccessProviderType("huaweicloud")
	AccessProviderTypeJDCloud             = AccessProviderType("jdcloud")
	AccessProviderTypeKong                = AccessProviderType("kong")
	AccessProviderTypeKubernetes          = AccessProviderType("k8s")
	AccessProviderTypeLarkBot             = AccessProviderType("larkbot")
	AccessProviderTypeLetsEncrypt         = AccessProviderType("letsencrypt")
	AccessProviderTypeLetsEncryptStaging  = AccessProviderType("letsencryptstaging")
	AccessProviderTypeLeCDN               = AccessProviderType("lecdn")
	AccessProviderTypeLocal               = AccessProviderType("local")
	AccessProviderTypeMattermost          = AccessProviderType("mattermost")
	AccessProviderTypeNamecheap           = AccessProviderType("namecheap")
	AccessProviderTypeNameDotCom          = AccessProviderType("namedotcom")
	AccessProviderTypeNameSilo            = AccessProviderType("namesilo")
	AccessProviderTypeNetcup              = AccessProviderType("netcup")
	AccessProviderTypeNetlify             = AccessProviderType("netlify")
	AccessProviderTypeNS1                 = AccessProviderType("ns1")
	AccessProviderTypePorkbun             = AccessProviderType("porkbun")
	AccessProviderTypePowerDNS            = AccessProviderType("powerdns")
	AccessProviderTypeProxmoxVE           = AccessProviderType("proxmoxve")
	AccessProviderTypeQiniu               = AccessProviderType("qiniu")
	AccessProviderTypeQingCloud           = AccessProviderType("qingcloud") // 青云（预留）
	AccessProviderTypeRainYun             = AccessProviderType("rainyun")
	AccessProviderTypeRatPanel            = AccessProviderType("ratpanel")
	AccessProviderTypeSafeLine            = AccessProviderType("safeline")
	AccessProviderTypeSlackBot            = AccessProviderType("slackbot")
	AccessProviderTypeSpaceship           = AccessProviderType("spaceship")
	AccessProviderTypeSSH                 = AccessProviderType("ssh")
	AccessProviderTypeSSLCOM              = AccessProviderType("sslcom")
	AccessProviderTypeTelegramBot         = AccessProviderType("telegrambot")
	AccessProviderTypeTencentCloud        = AccessProviderType("tencentcloud")
	AccessProviderTypeUCloud              = AccessProviderType("ucloud")
	AccessProviderTypeUniCloud            = AccessProviderType("unicloud")
	AccessProviderTypeUpyun               = AccessProviderType("upyun")
	AccessProviderTypeVercel              = AccessProviderType("vercel")
	AccessProviderTypeVolcEngine          = AccessProviderType("volcengine")
	AccessProviderTypeWangsu              = AccessProviderType("wangsu")
	AccessProviderTypeWebhook             = AccessProviderType("webhook")
	AccessProviderTypeWeComBot            = AccessProviderType("wecombot")
	AccessProviderTypeWestcn              = AccessProviderType("westcn")
	AccessProviderTypeZeroSSL             = AccessProviderType("zerossl")
)

type CAProviderType string

/*
证书颁发机构提供商常量值。
短横线前的部分始终等于授权提供商类型。

	注意：如果追加新的常量值，请保持以 ASCII 排序。
	NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	CAProviderTypeACMECA              = CAProviderType(AccessProviderTypeACMECA)
	CAProviderTypeBuypass             = CAProviderType(AccessProviderTypeBuypass)
	CAProviderTypeGoogleTrustServices = CAProviderType(AccessProviderTypeGoogleTrustServices)
	CAProviderTypeLetsEncrypt         = CAProviderType(AccessProviderTypeLetsEncrypt)
	CAProviderTypeLetsEncryptStaging  = CAProviderType(AccessProviderTypeLetsEncryptStaging)
	CAProviderTypeSSLCom              = CAProviderType(AccessProviderTypeSSLCOM)
	CAProviderTypeZeroSSL             = CAProviderType(AccessProviderTypeZeroSSL)
)

type ACMEDns01ProviderType string

/*
ACME DNS-01 提供商常量值。
短横线前的部分始终等于授权提供商类型。

	注意：如果追加新的常量值，请保持以 ASCII 排序。
	NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	ACMEDns01ProviderTypeACMEHttpReq       = ACMEDns01ProviderType(AccessProviderTypeACMEHttpReq)
	ACMEDns01ProviderTypeAliyun            = ACMEDns01ProviderType(AccessProviderTypeAliyun) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAliyunDNS]
	ACMEDns01ProviderTypeAliyunDNS         = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-dns")
	ACMEDns01ProviderTypeAliyunESA         = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-esa")
	ACMEDns01ProviderTypeAWS               = ACMEDns01ProviderType(AccessProviderTypeAWS) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAWSRoute53]
	ACMEDns01ProviderTypeAWSRoute53        = ACMEDns01ProviderType(AccessProviderTypeAWS + "-route53")
	ACMEDns01ProviderTypeAzure             = ACMEDns01ProviderType(AccessProviderTypeAzure) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAzure]
	ACMEDns01ProviderTypeAzureDNS          = ACMEDns01ProviderType(AccessProviderTypeAzure + "-dns")
	ACMEDns01ProviderTypeBaiduCloud        = ACMEDns01ProviderType(AccessProviderTypeBaiduCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeBaiduCloudDNS]
	ACMEDns01ProviderTypeBaiduCloudDNS     = ACMEDns01ProviderType(AccessProviderTypeBaiduCloud + "-dns")
	ACMEDns01ProviderTypeBunny             = ACMEDns01ProviderType(AccessProviderTypeBunny)
	ACMEDns01ProviderTypeCloudflare        = ACMEDns01ProviderType(AccessProviderTypeCloudflare)
	ACMEDns01ProviderTypeClouDNS           = ACMEDns01ProviderType(AccessProviderTypeClouDNS)
	ACMEDns01ProviderTypeCMCCCloud         = ACMEDns01ProviderType(AccessProviderTypeCMCCCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeCMCCCloudDNS]
	ACMEDns01ProviderTypeCMCCCloudDNS      = ACMEDns01ProviderType(AccessProviderTypeCMCCCloud + "-dns")
	ACMEDns01ProviderTypeConstellix        = ACMEDns01ProviderType(AccessProviderTypeConstellix)
	ACMEDns01ProviderTypeCTCCCloud         = ACMEDns01ProviderType(AccessProviderTypeCTCCCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeCTCCCloudSmartDNS]
	ACMEDns01ProviderTypeCTCCCloudSmartDNS = ACMEDns01ProviderType(AccessProviderTypeCTCCCloud + "-smartdns")
	ACMEDns01ProviderTypeDeSEC             = ACMEDns01ProviderType(AccessProviderTypeDeSEC)
	ACMEDns01ProviderTypeDigitalOcean      = ACMEDns01ProviderType(AccessProviderTypeDigitalOcean)
	ACMEDns01ProviderTypeDNSLA             = ACMEDns01ProviderType(AccessProviderTypeDNSLA)
	ACMEDns01ProviderTypeDuckDNS           = ACMEDns01ProviderType(AccessProviderTypeDuckDNS)
	ACMEDns01ProviderTypeDynv6             = ACMEDns01ProviderType(AccessProviderTypeDynv6)
	ACMEDns01ProviderTypeGcore             = ACMEDns01ProviderType(AccessProviderTypeGcore)
	ACMEDns01ProviderTypeGname             = ACMEDns01ProviderType(AccessProviderTypeGname)
	ACMEDns01ProviderTypeGoDaddy           = ACMEDns01ProviderType(AccessProviderTypeGoDaddy)
	ACMEDns01ProviderTypeHetzner           = ACMEDns01ProviderType(AccessProviderTypeHetzner)
	ACMEDns01ProviderTypeHuaweiCloud       = ACMEDns01ProviderType(AccessProviderTypeHuaweiCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeHuaweiCloudDNS]
	ACMEDns01ProviderTypeHuaweiCloudDNS    = ACMEDns01ProviderType(AccessProviderTypeHuaweiCloud + "-dns")
	ACMEDns01ProviderTypeJDCloud           = ACMEDns01ProviderType(AccessProviderTypeJDCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeJDCloudDNS]
	ACMEDns01ProviderTypeJDCloudDNS        = ACMEDns01ProviderType(AccessProviderTypeJDCloud + "-dns")
	ACMEDns01ProviderTypeNamecheap         = ACMEDns01ProviderType(AccessProviderTypeNamecheap)
	ACMEDns01ProviderTypeNameDotCom        = ACMEDns01ProviderType(AccessProviderTypeNameDotCom)
	ACMEDns01ProviderTypeNameSilo          = ACMEDns01ProviderType(AccessProviderTypeNameSilo)
	ACMEDns01ProviderTypeNetcup            = ACMEDns01ProviderType(AccessProviderTypeNetcup)
	ACMEDns01ProviderTypeNetlify           = ACMEDns01ProviderType(AccessProviderTypeNetlify)
	ACMEDns01ProviderTypeNS1               = ACMEDns01ProviderType(AccessProviderTypeNS1)
	ACMEDns01ProviderTypePorkbun           = ACMEDns01ProviderType(AccessProviderTypePorkbun)
	ACMEDns01ProviderTypePowerDNS          = ACMEDns01ProviderType(AccessProviderTypePowerDNS)
	ACMEDns01ProviderTypeRainYun           = ACMEDns01ProviderType(AccessProviderTypeRainYun)
	ACMEDns01ProviderTypeSpaceship         = ACMEDns01ProviderType(AccessProviderTypeSpaceship)
	ACMEDns01ProviderTypeTencentCloud      = ACMEDns01ProviderType(AccessProviderTypeTencentCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeTencentCloudDNS]
	ACMEDns01ProviderTypeTencentCloudDNS   = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-dns")
	ACMEDns01ProviderTypeTencentCloudEO    = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-eo")
	ACMEDns01ProviderTypeUCloudUDNR        = ACMEDns01ProviderType(AccessProviderTypeUCloud + "-udnr")
	ACMEDns01ProviderTypeVercel            = ACMEDns01ProviderType(AccessProviderTypeVercel)
	ACMEDns01ProviderTypeVolcEngine        = ACMEDns01ProviderType(AccessProviderTypeVolcEngine) // 兼容旧值，等同于 [ACMEDns01ProviderTypeVolcEngineDNS]
	ACMEDns01ProviderTypeVolcEngineDNS     = ACMEDns01ProviderType(AccessProviderTypeVolcEngine + "-dns")
	ACMEDns01ProviderTypeWestcn            = ACMEDns01ProviderType(AccessProviderTypeWestcn)
)

type DeploymentProviderType string

/*
部署证书主机提供商常量值。
短横线前的部分始终等于授权提供商类型。

	注意：如果追加新的常量值，请保持以 ASCII 排序。
	NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	DeploymentProviderType1PanelConsole         = DeploymentProviderType(AccessProviderType1Panel + "-console")
	DeploymentProviderType1PanelSite            = DeploymentProviderType(AccessProviderType1Panel + "-site")
	DeploymentProviderTypeAliyunALB             = DeploymentProviderType(AccessProviderTypeAliyun + "-alb")
	DeploymentProviderTypeAliyunAPIGW           = DeploymentProviderType(AccessProviderTypeAliyun + "-apigw")
	DeploymentProviderTypeAliyunCAS             = DeploymentProviderType(AccessProviderTypeAliyun + "-cas")
	DeploymentProviderTypeAliyunCASDeploy       = DeploymentProviderType(AccessProviderTypeAliyun + "-casdeploy")
	DeploymentProviderTypeAliyunCDN             = DeploymentProviderType(AccessProviderTypeAliyun + "-cdn")
	DeploymentProviderTypeAliyunCLB             = DeploymentProviderType(AccessProviderTypeAliyun + "-clb")
	DeploymentProviderTypeAliyunDCDN            = DeploymentProviderType(AccessProviderTypeAliyun + "-dcdn")
	DeploymentProviderTypeAliyunDDoS            = DeploymentProviderType(AccessProviderTypeAliyun + "-ddos")
	DeploymentProviderTypeAliyunESA             = DeploymentProviderType(AccessProviderTypeAliyun + "-esa")
	DeploymentProviderTypeAliyunFC              = DeploymentProviderType(AccessProviderTypeAliyun + "-fc")
	DeploymentProviderTypeAliyunGA              = DeploymentProviderType(AccessProviderTypeAliyun + "-ga")
	DeploymentProviderTypeAliyunLive            = DeploymentProviderType(AccessProviderTypeAliyun + "-live")
	DeploymentProviderTypeAliyunNLB             = DeploymentProviderType(AccessProviderTypeAliyun + "-nlb")
	DeploymentProviderTypeAliyunOSS             = DeploymentProviderType(AccessProviderTypeAliyun + "-oss")
	DeploymentProviderTypeAliyunVOD             = DeploymentProviderType(AccessProviderTypeAliyun + "-vod")
	DeploymentProviderTypeAliyunWAF             = DeploymentProviderType(AccessProviderTypeAliyun + "-waf")
	DeploymentProviderTypeAPISIX                = DeploymentProviderType(AccessProviderTypeAPISIX)
	DeploymentProviderTypeAWSACM                = DeploymentProviderType(AccessProviderTypeAWS + "-acm")
	DeploymentProviderTypeAWSCloudFront         = DeploymentProviderType(AccessProviderTypeAWS + "-cloudfront")
	DeploymentProviderTypeAWSIAM                = DeploymentProviderType(AccessProviderTypeAWS + "-iam")
	DeploymentProviderTypeAzureKeyVault         = DeploymentProviderType(AccessProviderTypeAzure + "-keyvault")
	DeploymentProviderTypeBaiduCloudAppBLB      = DeploymentProviderType(AccessProviderTypeBaiduCloud + "-appblb")
	DeploymentProviderTypeBaiduCloudBLB         = DeploymentProviderType(AccessProviderTypeBaiduCloud + "-blb")
	DeploymentProviderTypeBaiduCloudCDN         = DeploymentProviderType(AccessProviderTypeBaiduCloud + "-cdn")
	DeploymentProviderTypeBaiduCloudCert        = DeploymentProviderType(AccessProviderTypeBaiduCloud + "-cert")
	DeploymentProviderTypeBaishanCDN            = DeploymentProviderType(AccessProviderTypeBaishan + "-cdn")
	DeploymentProviderTypeBaotaPanelConsole     = DeploymentProviderType(AccessProviderTypeBaotaPanel + "-console")
	DeploymentProviderTypeBaotaPanelSite        = DeploymentProviderType(AccessProviderTypeBaotaPanel + "-site")
	DeploymentProviderTypeBaotaWAFConsole       = DeploymentProviderType(AccessProviderTypeBaotaWAF + "-console")
	DeploymentProviderTypeBaotaWAFSite          = DeploymentProviderType(AccessProviderTypeBaotaWAF + "-site")
	DeploymentProviderTypeBunnyCDN              = DeploymentProviderType(AccessProviderTypeBunny + "-cdn")
	DeploymentProviderTypeBytePlusCDN           = DeploymentProviderType(AccessProviderTypeBytePlus + "-cdn")
	DeploymentProviderTypeCacheFly              = DeploymentProviderType(AccessProviderTypeCacheFly)
	DeploymentProviderTypeCdnfly                = DeploymentProviderType(AccessProviderTypeCdnfly)
	DeploymentProviderTypeCTCCCloudAO           = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-ao")
	DeploymentProviderTypeCTCCCloudCDN          = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-cdn")
	DeploymentProviderTypeCTCCCloudCMS          = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-cms")
	DeploymentProviderTypeCTCCCloudELB          = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-elb")
	DeploymentProviderTypeCTCCCloudICDN         = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-icdn")
	DeploymentProviderTypeCTCCCloudLVDN         = DeploymentProviderType(AccessProviderTypeCTCCCloud + "-ldvn")
	DeploymentProviderTypeDogeCloudCDN          = DeploymentProviderType(AccessProviderTypeDogeCloud + "-cdn")
	DeploymentProviderTypeEdgioApplications     = DeploymentProviderType(AccessProviderTypeEdgio + "-applications")
	DeploymentProviderTypeFlexCDN               = DeploymentProviderType(AccessProviderTypeFlexCDN)
	DeploymentProviderTypeGcoreCDN              = DeploymentProviderType(AccessProviderTypeGcore + "-cdn")
	DeploymentProviderTypeGoEdge                = DeploymentProviderType(AccessProviderTypeGoEdge)
	DeploymentProviderTypeHuaweiCloudCDN        = DeploymentProviderType(AccessProviderTypeHuaweiCloud + "-cdn")
	DeploymentProviderTypeHuaweiCloudELB        = DeploymentProviderType(AccessProviderTypeHuaweiCloud + "-elb")
	DeploymentProviderTypeHuaweiCloudSCM        = DeploymentProviderType(AccessProviderTypeHuaweiCloud + "-scm")
	DeploymentProviderTypeHuaweiCloudWAF        = DeploymentProviderType(AccessProviderTypeHuaweiCloud + "-waf")
	DeploymentProviderTypeJDCloudALB            = DeploymentProviderType(AccessProviderTypeJDCloud + "-alb")
	DeploymentProviderTypeJDCloudCDN            = DeploymentProviderType(AccessProviderTypeJDCloud + "-cdn")
	DeploymentProviderTypeJDCloudLive           = DeploymentProviderType(AccessProviderTypeJDCloud + "-live")
	DeploymentProviderTypeJDCloudVOD            = DeploymentProviderType(AccessProviderTypeJDCloud + "-vod")
	DeploymentProviderTypeKong                  = DeploymentProviderType(AccessProviderTypeKong)
	DeploymentProviderTypeKubernetesSecret      = DeploymentProviderType(AccessProviderTypeKubernetes + "-secret")
	DeploymentProviderTypeLeCDN                 = DeploymentProviderType(AccessProviderTypeLeCDN)
	DeploymentProviderTypeLocal                 = DeploymentProviderType(AccessProviderTypeLocal)
	DeploymentProviderTypeNetlifySite           = DeploymentProviderType(AccessProviderTypeNetlify + "-site")
	DeploymentProviderTypeProxmoxVE             = DeploymentProviderType(AccessProviderTypeProxmoxVE)
	DeploymentProviderTypeQiniuCDN              = DeploymentProviderType(AccessProviderTypeQiniu + "-cdn")
	DeploymentProviderTypeQiniuKodo             = DeploymentProviderType(AccessProviderTypeQiniu + "-kodo")
	DeploymentProviderTypeQiniuPili             = DeploymentProviderType(AccessProviderTypeQiniu + "-pili")
	DeploymentProviderTypeRainYunRCDN           = DeploymentProviderType(AccessProviderTypeRainYun + "-rcdn")
	DeploymentProviderTypeRatPanelConsole       = DeploymentProviderType(AccessProviderTypeRatPanel + "-console")
	DeploymentProviderTypeRatPanelSite          = DeploymentProviderType(AccessProviderTypeRatPanel + "-site")
	DeploymentProviderTypeSafeLine              = DeploymentProviderType(AccessProviderTypeSafeLine)
	DeploymentProviderTypeSSH                   = DeploymentProviderType(AccessProviderTypeSSH)
	DeploymentProviderTypeTencentCloudCDN       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-cdn")
	DeploymentProviderTypeTencentCloudCLB       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-clb")
	DeploymentProviderTypeTencentCloudCOS       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-cos")
	DeploymentProviderTypeTencentCloudCSS       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-css")
	DeploymentProviderTypeTencentCloudECDN      = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ecdn")
	DeploymentProviderTypeTencentCloudEO        = DeploymentProviderType(AccessProviderTypeTencentCloud + "-eo")
	DeploymentProviderTypeTencentCloudGAAP      = DeploymentProviderType(AccessProviderTypeTencentCloud + "-gaap")
	DeploymentProviderTypeTencentCloudSCF       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-scf")
	DeploymentProviderTypeTencentCloudSSL       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ssl")
	DeploymentProviderTypeTencentCloudSSLDeploy = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ssldeploy")
	DeploymentProviderTypeTencentCloudSSLUpdate = DeploymentProviderType(AccessProviderTypeTencentCloud + "-sslupdate")
	DeploymentProviderTypeTencentCloudVOD       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-vod")
	DeploymentProviderTypeTencentCloudWAF       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-waf")
	DeploymentProviderTypeUCloudUCDN            = DeploymentProviderType(AccessProviderTypeUCloud + "-ucdn")
	DeploymentProviderTypeUCloudUS3             = DeploymentProviderType(AccessProviderTypeUCloud + "-us3")
	DeploymentProviderTypeUniCloudWebHost       = DeploymentProviderType(AccessProviderTypeUniCloud + "-webhost")
	DeploymentProviderTypeUpyunCDN              = DeploymentProviderType(AccessProviderTypeUpyun + "-cdn")
	DeploymentProviderTypeUpyunFile             = DeploymentProviderType(AccessProviderTypeUpyun + "-file")
	DeploymentProviderTypeVolcEngineALB         = DeploymentProviderType(AccessProviderTypeVolcEngine + "-alb")
	DeploymentProviderTypeVolcEngineCDN         = DeploymentProviderType(AccessProviderTypeVolcEngine + "-cdn")
	DeploymentProviderTypeVolcEngineCertCenter  = DeploymentProviderType(AccessProviderTypeVolcEngine + "-certcenter")
	DeploymentProviderTypeVolcEngineCLB         = DeploymentProviderType(AccessProviderTypeVolcEngine + "-clb")
	DeploymentProviderTypeVolcEngineDCDN        = DeploymentProviderType(AccessProviderTypeVolcEngine + "-dcdn")
	DeploymentProviderTypeVolcEngineImageX      = DeploymentProviderType(AccessProviderTypeVolcEngine + "-imagex")
	DeploymentProviderTypeVolcEngineLive        = DeploymentProviderType(AccessProviderTypeVolcEngine + "-live")
	DeploymentProviderTypeVolcEngineTOS         = DeploymentProviderType(AccessProviderTypeVolcEngine + "-tos")
	DeploymentProviderTypeWangsuCDN             = DeploymentProviderType(AccessProviderTypeWangsu + "-cdn")
	DeploymentProviderTypeWangsuCDNPro          = DeploymentProviderType(AccessProviderTypeWangsu + "-cdnpro")
	DeploymentProviderTypeWangsuCertificate     = DeploymentProviderType(AccessProviderTypeWangsu + "-certificate")
	DeploymentProviderTypeWebhook               = DeploymentProviderType(AccessProviderTypeWebhook)
)

type NotificationProviderType string

/*
消息通知提供商常量值。
短横线前的部分始终等于授权提供商类型。

	注意：如果追加新的常量值，请保持以 ASCII 排序。
	NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	NotificationProviderTypeDingTalkBot = NotificationProviderType(AccessProviderTypeDingTalkBot)
	NotificationProviderTypeDiscordBot  = NotificationProviderType(AccessProviderTypeDiscordBot)
	NotificationProviderTypeEmail       = NotificationProviderType(AccessProviderTypeEmail)
	NotificationProviderTypeLarkBot     = NotificationProviderType(AccessProviderTypeLarkBot)
	NotificationProviderTypeMattermost  = NotificationProviderType(AccessProviderTypeMattermost)
	NotificationProviderTypeSlackBot    = NotificationProviderType(AccessProviderTypeSlackBot)
	NotificationProviderTypeTelegramBot = NotificationProviderType(AccessProviderTypeTelegramBot)
	NotificationProviderTypeWebhook     = NotificationProviderType(AccessProviderTypeWebhook)
	NotificationProviderTypeWeComBot    = NotificationProviderType(AccessProviderTypeWeComBot)
)
