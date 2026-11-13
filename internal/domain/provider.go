package domain

type AccessProviderType string

func (t AccessProviderType) String() string {
	return string(t)
}

/*
授权提供商类型常量值。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	AccessProviderType35cn                = AccessProviderType("35cn")
	AccessProviderType51DNScom            = AccessProviderType("51dnscom")
	AccessProviderTypeACMECA              = AccessProviderType("acmeca")
	AccessProviderTypeACMEDNS             = AccessProviderType("acmedns")
	AccessProviderTypeACMEHttpReq         = AccessProviderType("acmehttpreq")
	AccessProviderTypeActalisSSL          = AccessProviderType("actalisssl")
	AccessProviderTypeAkamai              = AccessProviderType("akamai")
	AccessProviderTypeAliyun              = AccessProviderType("aliyun")
	AccessProviderTypeArvanCloud          = AccessProviderType("arvancloud")
	AccessProviderTypeAWS                 = AccessProviderType("aws")
	AccessProviderTypeAzure               = AccessProviderType("azure")
	AccessProviderTypeBaiduCloud          = AccessProviderType("baiducloud")
	AccessProviderTypeBeget               = AccessProviderType("beget")
	AccessProviderTypeBookMyName          = AccessProviderType("bookmyname")
	AccessProviderTypeBunny               = AccessProviderType("bunny")
	AccessProviderTypeBytePlus            = AccessProviderType("byteplus")
	AccessProviderTypeCloudflare          = AccessProviderType("cloudflare")
	AccessProviderTypeClouDNS             = AccessProviderType("cloudns")
	AccessProviderTypeCMCCCloud           = AccessProviderType("cmcccloud")
	AccessProviderTypeConoHaVPS           = AccessProviderType("conohavps")
	AccessProviderTypeConstellix          = AccessProviderType("constellix")
	AccessProviderTypeCPanel              = AccessProviderType("cpanel")
	AccessProviderTypeCTCCCloud           = AccessProviderType("ctcccloud")
	AccessProviderTypeCUCCCloud           = AccessProviderType("cucccloud") // 联通云（预留）
	AccessProviderTypeDeSEC               = AccessProviderType("desec")
	AccessProviderTypeDigiCert            = AccessProviderType("digicert")
	AccessProviderTypeDigitalOcean        = AccessProviderType("digitalocean")
	AccessProviderTypeDNSExit             = AccessProviderType("dnsexit")
	AccessProviderTypeDNSLA               = AccessProviderType("dnsla")
	AccessProviderTypeDNSMadeEasy         = AccessProviderType("dnsmadeeasy")
	AccessProviderTypeDuckDNS             = AccessProviderType("duckdns")
	AccessProviderTypeDynadot             = AccessProviderType("dynadot")
	AccessProviderTypeDynu                = AccessProviderType("dynu")
	AccessProviderTypeDynv6               = AccessProviderType("dynv6")
	AccessProviderTypeEmail               = AccessProviderType("email")
	AccessProviderTypeFastly              = AccessProviderType("fastly") // Fastly（预留）
	AccessProviderTypeFTP                 = AccessProviderType("ftp")
	AccessProviderTypeGandinet            = AccessProviderType("gandinet")
	AccessProviderTypeGcore               = AccessProviderType("gcore")
	AccessProviderTypeGlobalSignAtlas     = AccessProviderType("globalsignatlas")
	AccessProviderTypeGname               = AccessProviderType("gname")
	AccessProviderTypeGoDaddy             = AccessProviderType("godaddy")
	AccessProviderTypeGoogleCloud         = AccessProviderType("googlecloud")
	AccessProviderTypeGoogleTrustServices = AccessProviderType("googletrustservices")
	AccessProviderTypeHetzner             = AccessProviderType("hetzner")
	AccessProviderTypeHostingde           = AccessProviderType("hostingde")
	AccessProviderTypeHostinger           = AccessProviderType("hostinger")
	AccessProviderTypeHuaweiCloud         = AccessProviderType("huaweicloud")
	AccessProviderTypeInfomaniak          = AccessProviderType("infomaniak")
	AccessProviderTypeIONOS               = AccessProviderType("ionos")
	AccessProviderTypeJDCloud             = AccessProviderType("jdcloud")
	AccessProviderTypeLetsEncrypt         = AccessProviderType("letsencrypt")
	AccessProviderTypeLetsEncryptStaging  = AccessProviderType("letsencryptstaging")
	AccessProviderTypeLinode              = AccessProviderType("linode")
	AccessProviderTypeLiteSSL             = AccessProviderType("litessl")
	AccessProviderTypeLocal               = AccessProviderType("local")
	AccessProviderTypeLuaDNS              = AccessProviderType("luadns")
	AccessProviderTypeNamecheap           = AccessProviderType("namecheap")
	AccessProviderTypeNameDotCom          = AccessProviderType("namedotcom")
	AccessProviderTypeNameSilo            = AccessProviderType("namesilo")
	AccessProviderTypeNetcup              = AccessProviderType("netcup")
	AccessProviderTypeNetlify             = AccessProviderType("netlify")
	AccessProviderTypeNS1                 = AccessProviderType("ns1")
	AccessProviderTypeOracleCloud         = AccessProviderType("oraclecloud")
	AccessProviderTypeOVHcloud            = AccessProviderType("ovhcloud")
	AccessProviderTypePorkbun             = AccessProviderType("porkbun")
	AccessProviderTypePowerDNS            = AccessProviderType("powerdns")
	AccessProviderTypeQingCloud           = AccessProviderType("qingcloud")
	AccessProviderTypeRainYun             = AccessProviderType("rainyun")
	AccessProviderTypeRegru               = AccessProviderType("regru")
	AccessProviderTypeRFC2136             = AccessProviderType("rfc2136")
	AccessProviderTypeRuCenter            = AccessProviderType("rucenter")
	AccessProviderTypeS3                  = AccessProviderType("s3")
	AccessProviderTypeSectigo             = AccessProviderType("sectigo")
	AccessProviderTypeSimplyCom           = AccessProviderType("simplycom")
	AccessProviderTypeSpaceship           = AccessProviderType("spaceship")
	AccessProviderTypeSSH                 = AccessProviderType("ssh")
	AccessProviderTypeSSLCOM              = AccessProviderType("sslcom")
	AccessProviderTypeTechnitiumDNS       = AccessProviderType("technitiumdns")
	AccessProviderTypeTencentCloud        = AccessProviderType("tencentcloud")
	AccessProviderTypeTodayNIC            = AccessProviderType("todaynic")
	AccessProviderTypeUCloud              = AccessProviderType("ucloud")
	AccessProviderTypeVercel              = AccessProviderType("vercel")
	AccessProviderTypeVolcEngine          = AccessProviderType("volcengine")
	AccessProviderTypeVultr               = AccessProviderType("vultr")
	AccessProviderTypeWebhook             = AccessProviderType("webhook")
	AccessProviderTypeWestcn              = AccessProviderType("westcn")
	AccessProviderTypeXinnet              = AccessProviderType("xinnet")
	AccessProviderTypeYandexCloud         = AccessProviderType("yandexcloud")
	AccessProviderTypeZeroSSL             = AccessProviderType("zerossl")
)

type CAProviderType string

func (t CAProviderType) String() string {
	return string(t)
}

/*
证书颁发机构提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	CAProviderTypeACMECA              = CAProviderType(AccessProviderTypeACMECA)
	CAProviderTypeActalisSSL          = CAProviderType(AccessProviderTypeActalisSSL)
	CAProviderTypeDigiCert            = CAProviderType(AccessProviderTypeDigiCert)
	CAProviderTypeGlobalSignAtlas     = CAProviderType(AccessProviderTypeGlobalSignAtlas)
	CAProviderTypeGoogleTrustServices = CAProviderType(AccessProviderTypeGoogleTrustServices)
	CAProviderTypeLetsEncrypt         = CAProviderType(AccessProviderTypeLetsEncrypt)
	CAProviderTypeLetsEncryptStaging  = CAProviderType(AccessProviderTypeLetsEncryptStaging)
	CAProviderTypeLiteSSL             = CAProviderType(AccessProviderTypeLiteSSL)
	CAProviderTypeSectigo             = CAProviderType(AccessProviderTypeSectigo)
	CAProviderTypeSSLCom              = CAProviderType(AccessProviderTypeSSLCOM)
	CAProviderTypeZeroSSL             = CAProviderType(AccessProviderTypeZeroSSL)
)

type ACMEChallengeProviderType string

func (t ACMEChallengeProviderType) String() string {
	return string(t)
}

type ACMEDns01ProviderType ACMEChallengeProviderType

func (t ACMEDns01ProviderType) String() string {
	return string(t)
}

/*
ACME DNS-01 提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	ACMEDns01ProviderType35cn              = ACMEDns01ProviderType(AccessProviderType35cn)
	ACMEDns01ProviderType51DNScom          = ACMEDns01ProviderType(AccessProviderType51DNScom)
	ACMEDns01ProviderTypeACMEDNS           = ACMEDns01ProviderType(AccessProviderTypeACMEDNS)
	ACMEDns01ProviderTypeACMEHttpReq       = ACMEDns01ProviderType(AccessProviderTypeACMEHttpReq)
	ACMEDns01ProviderTypeAkamai            = ACMEDns01ProviderType(AccessProviderTypeAkamai) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAkamaiEdgeDNS]
	ACMEDns01ProviderTypeAkamaiEdgeDNS     = ACMEDns01ProviderType(AccessProviderTypeAkamai + "-edgedns")
	ACMEDns01ProviderTypeAliyun            = ACMEDns01ProviderType(AccessProviderTypeAliyun) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAliyunDNS]
	ACMEDns01ProviderTypeAliyunDNS         = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-dns")
	ACMEDns01ProviderTypeAliyunESA         = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-esa")
	ACMEDns01ProviderTypeArvanCloud        = ACMEDns01ProviderType(AccessProviderTypeArvanCloud)
	ACMEDns01ProviderTypeAWS               = ACMEDns01ProviderType(AccessProviderTypeAWS) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAWSRoute53]
	ACMEDns01ProviderTypeAWSLightsail      = ACMEDns01ProviderType(AccessProviderTypeAWS + "-lightsail")
	ACMEDns01ProviderTypeAWSRoute53        = ACMEDns01ProviderType(AccessProviderTypeAWS + "-route53")
	ACMEDns01ProviderTypeAzure             = ACMEDns01ProviderType(AccessProviderTypeAzure) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAzureDNS]
	ACMEDns01ProviderTypeAzureDNS          = ACMEDns01ProviderType(AccessProviderTypeAzure + "-dns")
	ACMEDns01ProviderTypeBaiduCloud        = ACMEDns01ProviderType(AccessProviderTypeBaiduCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeBaiduCloudDNS]
	ACMEDns01ProviderTypeBaiduCloudDNS     = ACMEDns01ProviderType(AccessProviderTypeBaiduCloud + "-dns")
	ACMEDns01ProviderTypeBeget             = ACMEDns01ProviderType(AccessProviderTypeBeget)
	ACMEDns01ProviderTypeBookMyName        = ACMEDns01ProviderType(AccessProviderTypeBookMyName)
	ACMEDns01ProviderTypeBunny             = ACMEDns01ProviderType(AccessProviderTypeBunny)
	ACMEDns01ProviderTypeBytePlus          = ACMEDns01ProviderType(AccessProviderTypeBytePlus) // 兼容旧值，等同于 [ACMEDns01ProviderTypeBytePlusDNS]
	ACMEDns01ProviderTypeBytePlusDNS       = ACMEDns01ProviderType(AccessProviderTypeBytePlus + "-dns")
	ACMEDns01ProviderTypeCloudflare        = ACMEDns01ProviderType(AccessProviderTypeCloudflare)
	ACMEDns01ProviderTypeClouDNS           = ACMEDns01ProviderType(AccessProviderTypeClouDNS)
	ACMEDns01ProviderTypeCMCCCloud         = ACMEDns01ProviderType(AccessProviderTypeCMCCCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeCMCCCloudDNS]
	ACMEDns01ProviderTypeCMCCCloudDNS      = ACMEDns01ProviderType(AccessProviderTypeCMCCCloud + "-dns")
	ACMEDns01ProviderTypeConoHaVPS         = ACMEDns01ProviderType(AccessProviderTypeConoHaVPS)
	ACMEDns01ProviderTypeConstellix        = ACMEDns01ProviderType(AccessProviderTypeConstellix)
	ACMEDns01ProviderTypeCPanel            = ACMEDns01ProviderType(AccessProviderTypeCPanel)
	ACMEDns01ProviderTypeCTCCCloud         = ACMEDns01ProviderType(AccessProviderTypeCTCCCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeCTCCCloudSmartDNS]
	ACMEDns01ProviderTypeCTCCCloudSmartDNS = ACMEDns01ProviderType(AccessProviderTypeCTCCCloud + "-smartdns")
	ACMEDns01ProviderTypeDeSEC             = ACMEDns01ProviderType(AccessProviderTypeDeSEC)
	ACMEDns01ProviderTypeDigitalOcean      = ACMEDns01ProviderType(AccessProviderTypeDigitalOcean)
	ACMEDns01ProviderTypeDNSExit           = ACMEDns01ProviderType(AccessProviderTypeDNSExit)
	ACMEDns01ProviderTypeDNSLA             = ACMEDns01ProviderType(AccessProviderTypeDNSLA)
	ACMEDns01ProviderTypeDNSMadeEasy       = ACMEDns01ProviderType(AccessProviderTypeDNSMadeEasy)
	ACMEDns01ProviderTypeDuckDNS           = ACMEDns01ProviderType(AccessProviderTypeDuckDNS)
	ACMEDns01ProviderTypeDynadot           = ACMEDns01ProviderType(AccessProviderTypeDynadot)
	ACMEDns01ProviderTypeDynu              = ACMEDns01ProviderType(AccessProviderTypeDynu)
	ACMEDns01ProviderTypeDynv6             = ACMEDns01ProviderType(AccessProviderTypeDynv6)
	ACMEDns01ProviderTypeGandinet          = ACMEDns01ProviderType(AccessProviderTypeGandinet)
	ACMEDns01ProviderTypeGcore             = ACMEDns01ProviderType(AccessProviderTypeGcore)
	ACMEDns01ProviderTypeGname             = ACMEDns01ProviderType(AccessProviderTypeGname)
	ACMEDns01ProviderTypeGoDaddy           = ACMEDns01ProviderType(AccessProviderTypeGoDaddy)
	ACMEDns01ProviderTypeGoogleCloud       = ACMEDns01ProviderType(AccessProviderTypeGoogleCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeGoogleCloudDNS]
	ACMEDns01ProviderTypeGoogleCloudDNS    = ACMEDns01ProviderType(AccessProviderTypeGoogleCloud + "-dns")
	ACMEDns01ProviderTypeHetzner           = ACMEDns01ProviderType(AccessProviderTypeHetzner)
	ACMEDns01ProviderTypeHostingde         = ACMEDns01ProviderType(AccessProviderTypeHostingde)
	ACMEDns01ProviderTypeHostinger         = ACMEDns01ProviderType(AccessProviderTypeHostinger)
	ACMEDns01ProviderTypeHuaweiCloud       = ACMEDns01ProviderType(AccessProviderTypeHuaweiCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeHuaweiCloudDNS]
	ACMEDns01ProviderTypeHuaweiCloudDNS    = ACMEDns01ProviderType(AccessProviderTypeHuaweiCloud + "-dns")
	ACMEDns01ProviderTypeInfomaniak        = ACMEDns01ProviderType(AccessProviderTypeInfomaniak)
	ACMEDns01ProviderTypeIONOS             = ACMEDns01ProviderType(AccessProviderTypeIONOS)
	ACMEDns01ProviderTypeJDCloud           = ACMEDns01ProviderType(AccessProviderTypeJDCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeJDCloudDNS]
	ACMEDns01ProviderTypeJDCloudDNS        = ACMEDns01ProviderType(AccessProviderTypeJDCloud + "-dns")
	ACMEDns01ProviderTypeLinode            = ACMEDns01ProviderType(AccessProviderTypeLinode)
	ACMEDns01ProviderTypeLuaDNS            = ACMEDns01ProviderType(AccessProviderTypeLuaDNS)
	ACMEDns01ProviderTypeNamecheap         = ACMEDns01ProviderType(AccessProviderTypeNamecheap)
	ACMEDns01ProviderTypeNameDotCom        = ACMEDns01ProviderType(AccessProviderTypeNameDotCom)
	ACMEDns01ProviderTypeNameSilo          = ACMEDns01ProviderType(AccessProviderTypeNameSilo)
	ACMEDns01ProviderTypeNetcup            = ACMEDns01ProviderType(AccessProviderTypeNetcup)
	ACMEDns01ProviderTypeNetlify           = ACMEDns01ProviderType(AccessProviderTypeNetlify)
	ACMEDns01ProviderTypeNS1               = ACMEDns01ProviderType(AccessProviderTypeNS1)
	ACMEDns01ProviderTypeOracleCloud       = ACMEDns01ProviderType(AccessProviderTypeOracleCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeOracleCloudDNS]
	ACMEDns01ProviderTypeOracleCloudDNS    = ACMEDns01ProviderType(AccessProviderTypeOracleCloud + "-dns")
	ACMEDns01ProviderTypeOVHcloud          = ACMEDns01ProviderType(AccessProviderTypeOVHcloud)
	ACMEDns01ProviderTypePorkbun           = ACMEDns01ProviderType(AccessProviderTypePorkbun)
	ACMEDns01ProviderTypePowerDNS          = ACMEDns01ProviderType(AccessProviderTypePowerDNS)
	ACMEDns01ProviderTypeQingCloud         = ACMEDns01ProviderType(AccessProviderTypeQingCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeQingCloudDNS]
	ACMEDns01ProviderTypeQingCloudDNS      = ACMEDns01ProviderType(AccessProviderTypeQingCloud + "-dns")
	ACMEDns01ProviderTypeRainYun           = ACMEDns01ProviderType(AccessProviderTypeRainYun)
	ACMEDns01ProviderTypeRegru             = ACMEDns01ProviderType(AccessProviderTypeRegru)
	ACMEDns01ProviderTypeRFC2136           = ACMEDns01ProviderType(AccessProviderTypeRFC2136)
	ACMEDns01ProviderTypeRuCenter          = ACMEDns01ProviderType(AccessProviderTypeRuCenter)
	ACMEDns01ProviderTypeSimplyCom         = ACMEDns01ProviderType(AccessProviderTypeSimplyCom)
	ACMEDns01ProviderTypeSpaceship         = ACMEDns01ProviderType(AccessProviderTypeSpaceship)
	ACMEDns01ProviderTypeTechnitiumDNS     = ACMEDns01ProviderType(AccessProviderTypeTechnitiumDNS)
	ACMEDns01ProviderTypeTencentCloud      = ACMEDns01ProviderType(AccessProviderTypeTencentCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeTencentCloudDNS]
	ACMEDns01ProviderTypeTencentCloudDNS   = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-dns")
	ACMEDns01ProviderTypeTencentCloudEO    = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-eo")
	ACMEDns01ProviderTypeTodayNIC          = ACMEDns01ProviderType(AccessProviderTypeTodayNIC)
	ACMEDns01ProviderTypeUCloud            = ACMEDns01ProviderType(AccessProviderTypeUCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeUCloudUDNR]
	ACMEDns01ProviderTypeUCloudUDNR        = ACMEDns01ProviderType(AccessProviderTypeUCloud + "-udnr")
	ACMEDns01ProviderTypeVercel            = ACMEDns01ProviderType(AccessProviderTypeVercel)
	ACMEDns01ProviderTypeVolcEngine        = ACMEDns01ProviderType(AccessProviderTypeVolcEngine) // 兼容旧值，等同于 [ACMEDns01ProviderTypeVolcEngineDNS]
	ACMEDns01ProviderTypeVolcEngineDNS     = ACMEDns01ProviderType(AccessProviderTypeVolcEngine + "-dns")
	ACMEDns01ProviderTypeVultr             = ACMEDns01ProviderType(AccessProviderTypeVultr)
	ACMEDns01ProviderTypeWestcn            = ACMEDns01ProviderType(AccessProviderTypeWestcn)
	ACMEDns01ProviderTypeXinnet            = ACMEDns01ProviderType(AccessProviderTypeXinnet)
	ACMEDns01ProviderTypeYandexCloud       = ACMEDns01ProviderType(AccessProviderTypeYandexCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeYandexCloudDNS]
	ACMEDns01ProviderTypeYandexCloudDNS    = ACMEDns01ProviderType(AccessProviderTypeYandexCloud + "-dns")
)

type ACMEHttp01ProviderType ACMEChallengeProviderType

func (t ACMEHttp01ProviderType) String() string {
	return t.String()
}

/*
ACME HTTP-01 提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	ACMEHttp01ProviderTypeLocal = ACMEHttp01ProviderType(AccessProviderTypeLocal)
	ACMEHttp01ProviderTypeFTP   = ACMEHttp01ProviderType(AccessProviderTypeFTP)
	ACMEHttp01ProviderTypeS3    = ACMEHttp01ProviderType(AccessProviderTypeS3)
	ACMEHttp01ProviderTypeSSH   = ACMEHttp01ProviderType(AccessProviderTypeSSH)
)

type DeploymentProviderType string

func (t DeploymentProviderType) String() string {
	return string(t)
}

/*
部署证书主机提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	DeploymentProviderTypeAliyunALB             = DeploymentProviderType(AccessProviderTypeAliyun + "-alb")
	DeploymentProviderTypeAliyunAPIGW           = DeploymentProviderType(AccessProviderTypeAliyun + "-apigw")
	DeploymentProviderTypeAliyunCAS             = DeploymentProviderType(AccessProviderTypeAliyun + "-cas")
	DeploymentProviderTypeAliyunCASDeploy       = DeploymentProviderType(AccessProviderTypeAliyun + "-casdeploy")
	DeploymentProviderTypeAliyunCDN             = DeploymentProviderType(AccessProviderTypeAliyun + "-cdn")
	DeploymentProviderTypeAliyunCLB             = DeploymentProviderType(AccessProviderTypeAliyun + "-clb")
	DeploymentProviderTypeAliyunDCDN            = DeploymentProviderType(AccessProviderTypeAliyun + "-dcdn")
	DeploymentProviderTypeAliyunDDoSPro         = DeploymentProviderType(AccessProviderTypeAliyun + "-ddospro")
	DeploymentProviderTypeAliyunESA             = DeploymentProviderType(AccessProviderTypeAliyun + "-esa")
	DeploymentProviderTypeAliyunESASaaS         = DeploymentProviderType(AccessProviderTypeAliyun + "-esasaas")
	DeploymentProviderTypeAliyunFC              = DeploymentProviderType(AccessProviderTypeAliyun + "-fc")
	DeploymentProviderTypeAliyunGA              = DeploymentProviderType(AccessProviderTypeAliyun + "-ga")
	DeploymentProviderTypeAliyunLive            = DeploymentProviderType(AccessProviderTypeAliyun + "-live")
	DeploymentProviderTypeAliyunMSE             = DeploymentProviderType(AccessProviderTypeAliyun + "-mse")
	DeploymentProviderTypeAliyunNLB             = DeploymentProviderType(AccessProviderTypeAliyun + "-nlb")
	DeploymentProviderTypeAliyunOSS             = DeploymentProviderType(AccessProviderTypeAliyun + "-oss")
	DeploymentProviderTypeAliyunVOD             = DeploymentProviderType(AccessProviderTypeAliyun + "-vod")
	DeploymentProviderTypeAliyunWAF             = DeploymentProviderType(AccessProviderTypeAliyun + "-waf")
	DeploymentProviderTypeAWSACM                = DeploymentProviderType(AccessProviderTypeAWS + "-acm")
	DeploymentProviderTypeAWSALB                = DeploymentProviderType(AccessProviderTypeAWS + "-alb")
	DeploymentProviderTypeAWSAmplify            = DeploymentProviderType(AccessProviderTypeAWS + "-amplify")
	DeploymentProviderTypeAWSAPIGateway         = DeploymentProviderType(AccessProviderTypeAWS + "-apigateway")
	DeploymentProviderTypeAWSCLB                = DeploymentProviderType(AccessProviderTypeAWS + "-clb")
	DeploymentProviderTypeAWSCloudFront         = DeploymentProviderType(AccessProviderTypeAWS + "-cloudfront")
	DeploymentProviderTypeAWSIAM                = DeploymentProviderType(AccessProviderTypeAWS + "-iam")
	DeploymentProviderTypeAWSNLB                = DeploymentProviderType(AccessProviderTypeAWS + "-nlb")
	DeploymentProviderTypeAzureKeyVault         = DeploymentProviderType(AccessProviderTypeAzure + "-keyvault")
	DeploymentProviderTypeFTP                   = DeploymentProviderType(AccessProviderTypeFTP)
	DeploymentProviderTypeLocal                 = DeploymentProviderType(AccessProviderTypeLocal)
	DeploymentProviderTypeS3                    = DeploymentProviderType(AccessProviderTypeS3)
	DeploymentProviderTypeSSH                   = DeploymentProviderType(AccessProviderTypeSSH)
	DeploymentProviderTypeTencentCloudCDN       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-cdn")
	DeploymentProviderTypeTencentCloudCLB       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-clb")
	DeploymentProviderTypeTencentCloudCOS       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-cos")
	DeploymentProviderTypeTencentCloudCSS       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-css")
	DeploymentProviderTypeTencentCloudDDoSPro   = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ddospro")
	DeploymentProviderTypeTencentCloudECDN      = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ecdn")
	DeploymentProviderTypeTencentCloudEO        = DeploymentProviderType(AccessProviderTypeTencentCloud + "-eo")
	DeploymentProviderTypeTencentCloudEOMakers  = DeploymentProviderType(AccessProviderTypeTencentCloud + "-eomakers")
	DeploymentProviderTypeTencentCloudGA2       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ga2")
	DeploymentProviderTypeTencentCloudGAAP      = DeploymentProviderType(AccessProviderTypeTencentCloud + "-gaap")
	DeploymentProviderTypeTencentCloudSCF       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-scf")
	DeploymentProviderTypeTencentCloudSSL       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ssl")
	DeploymentProviderTypeTencentCloudSSLDeploy = DeploymentProviderType(AccessProviderTypeTencentCloud + "-ssldeploy")
	DeploymentProviderTypeTencentCloudSSLUpdate = DeploymentProviderType(AccessProviderTypeTencentCloud + "-sslupdate")
	DeploymentProviderTypeTencentCloudTSE       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-tse")
	DeploymentProviderTypeTencentCloudVOD       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-vod")
	DeploymentProviderTypeTencentCloudWAF       = DeploymentProviderType(AccessProviderTypeTencentCloud + "-waf")
	DeploymentProviderTypeWebhook               = DeploymentProviderType(AccessProviderTypeWebhook)
)

type PurgeProviderType string

func (t PurgeProviderType) String() string {
	return string(t)
}

/*
清除证书主机提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	PurgeProviderTypeAliyunCAS            = PurgeProviderType(AccessProviderTypeAliyun + "-cas")
	PurgeProviderTypeAliyunCLB            = PurgeProviderType(AccessProviderTypeAliyun + "-clb")
	PurgeProviderTypeAliyunESA            = PurgeProviderType(AccessProviderTypeAliyun + "-esa")
	PurgeProviderTypeAWSACM               = PurgeProviderType(AccessProviderTypeAWS + "-acm")
	PurgeProviderTypeAWSIAM               = PurgeProviderType(AccessProviderTypeAWS + "-iam")
	PurgeProviderTypeAzureKeyVault        = PurgeProviderType(AccessProviderTypeAzure + "-keyvault")
	PurgeProviderTypeBaiduCloudCert       = PurgeProviderType(AccessProviderTypeBaiduCloud + "-cert")
	PurgeProviderTypeHuaweiCloudELB       = PurgeProviderType(AccessProviderTypeHuaweiCloud + "-elb")
	PurgeProviderTypeHuaweiCloudSCM       = PurgeProviderType(AccessProviderTypeHuaweiCloud + "-scm")
	PurgeProviderTypeHuaweiCloudWAF       = PurgeProviderType(AccessProviderTypeHuaweiCloud + "-waf")
	PurgeProviderTypeJDCloudSSL           = PurgeProviderType(AccessProviderTypeJDCloud + "-ssl")
	PurgeProviderTypeTencentCloudGAAP     = PurgeProviderType(AccessProviderTypeTencentCloud + "-gaap")
	PurgeProviderTypeTencentCloudSSL      = PurgeProviderType(AccessProviderTypeTencentCloud + "-ssl")
	PurgeProviderTypeTencentCloudTSE      = PurgeProviderType(AccessProviderTypeTencentCloud + "-tse")
	PurgeProviderTypeUCloudUSSL           = PurgeProviderType(AccessProviderTypeUCloud + "-ussl")
	PurgeProviderTypeVolcEngineCertCenter = PurgeProviderType(AccessProviderTypeVolcEngine + "-certcenter")
)

type NotificationProviderType string

func (t NotificationProviderType) String() string {
	return string(t)
}

/*
消息通知提供商常量值。
短横线前的部分始终等于授权提供商类型。

注意：如果追加新的常量值，请保持以 ASCII 排序。
NOTICE: If you add new constant, please keep ASCII order.
*/
const (
	NotificationProviderTypeEmail   = NotificationProviderType(AccessProviderTypeEmail)
	NotificationProviderTypeWebhook = NotificationProviderType(AccessProviderTypeWebhook)
)
