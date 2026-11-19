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
	AccessProviderTypeACMECA = AccessProviderType("acmeca")

	AccessProviderTypeActalisSSL = AccessProviderType("actalisssl")

	AccessProviderTypeAliyun = AccessProviderType("aliyun")

	AccessProviderTypeAWS        = AccessProviderType("aws")
	AccessProviderTypeAzure      = AccessProviderType("azure")
	AccessProviderTypeBaiduCloud = AccessProviderType("baiducloud")

	AccessProviderTypeBunny      = AccessProviderType("bunny")
	AccessProviderTypeBytePlus   = AccessProviderType("byteplus")
	AccessProviderTypeCloudflare = AccessProviderType("cloudflare")

	AccessProviderTypeCMCCCloud = AccessProviderType("cmcccloud")

	AccessProviderTypeCPanel    = AccessProviderType("cpanel")
	AccessProviderTypeCTCCCloud = AccessProviderType("ctcccloud")
	AccessProviderTypeCUCCCloud = AccessProviderType("cucccloud") // 联通云（预留）

	AccessProviderTypeDigiCert     = AccessProviderType("digicert")
	AccessProviderTypeDigitalOcean = AccessProviderType("digitalocean")

	AccessProviderTypeEmail  = AccessProviderType("email")
	AccessProviderTypeFastly = AccessProviderType("fastly") // Fastly（预留）
	AccessProviderTypeFTP    = AccessProviderType("ftp")

	AccessProviderTypeGcore           = AccessProviderType("gcore")
	AccessProviderTypeGlobalSignAtlas = AccessProviderType("globalsignatlas")

	AccessProviderTypeGoogleCloud         = AccessProviderType("googlecloud")
	AccessProviderTypeGoogleTrustServices = AccessProviderType("googletrustservices")

	AccessProviderTypeHuaweiCloud = AccessProviderType("huaweicloud")

	AccessProviderTypeJDCloud            = AccessProviderType("jdcloud")
	AccessProviderTypeLetsEncrypt        = AccessProviderType("letsencrypt")
	AccessProviderTypeLetsEncryptStaging = AccessProviderType("letsencryptstaging")
	AccessProviderTypeLinode             = AccessProviderType("linode")
	AccessProviderTypeLiteSSL            = AccessProviderType("litessl")
	AccessProviderTypeLocal              = AccessProviderType("local")

	AccessProviderTypeNetlify = AccessProviderType("netlify")

	AccessProviderTypeOracleCloud = AccessProviderType("oraclecloud")

	AccessProviderTypePowerDNS  = AccessProviderType("powerdns")
	AccessProviderTypeQingCloud = AccessProviderType("qingcloud")
	AccessProviderTypeRainYun   = AccessProviderType("rainyun")

	AccessProviderTypeRFC2136 = AccessProviderType("rfc2136")

	AccessProviderTypeS3      = AccessProviderType("s3")
	AccessProviderTypeSectigo = AccessProviderType("sectigo")

	AccessProviderTypeSSH    = AccessProviderType("ssh")
	AccessProviderTypeSSLCOM = AccessProviderType("sslcom")

	AccessProviderTypeTencentCloud = AccessProviderType("tencentcloud")

	AccessProviderTypeUCloud     = AccessProviderType("ucloud")
	AccessProviderTypeVercel     = AccessProviderType("vercel")
	AccessProviderTypeVolcEngine = AccessProviderType("volcengine")

	AccessProviderTypeWebhook = AccessProviderType("webhook")

	AccessProviderTypeYandexCloud = AccessProviderType("yandexcloud")
	AccessProviderTypeZeroSSL     = AccessProviderType("zerossl")
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
	ACMEDns01ProviderTypeAliyun          = ACMEDns01ProviderType(AccessProviderTypeAliyun) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAliyunDNS]
	ACMEDns01ProviderTypeAliyunDNS       = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-dns")
	ACMEDns01ProviderTypeAliyunESA       = ACMEDns01ProviderType(AccessProviderTypeAliyun + "-esa")
	ACMEDns01ProviderTypeAWS             = ACMEDns01ProviderType(AccessProviderTypeAWS) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAWSRoute53]
	ACMEDns01ProviderTypeAWSLightsail    = ACMEDns01ProviderType(AccessProviderTypeAWS + "-lightsail")
	ACMEDns01ProviderTypeAWSRoute53      = ACMEDns01ProviderType(AccessProviderTypeAWS + "-route53")
	ACMEDns01ProviderTypeAzure           = ACMEDns01ProviderType(AccessProviderTypeAzure) // 兼容旧值，等同于 [ACMEDns01ProviderTypeAzureDNS]
	ACMEDns01ProviderTypeAzureDNS        = ACMEDns01ProviderType(AccessProviderTypeAzure + "-dns")
	ACMEDns01ProviderTypeCloudflare      = ACMEDns01ProviderType(AccessProviderTypeCloudflare)
	ACMEDns01ProviderTypeTencentCloud    = ACMEDns01ProviderType(AccessProviderTypeTencentCloud) // 兼容旧值，等同于 [ACMEDns01ProviderTypeTencentCloudDNS]
	ACMEDns01ProviderTypeTencentCloudDNS = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-dns")
	ACMEDns01ProviderTypeTencentCloudEO  = ACMEDns01ProviderType(AccessProviderTypeTencentCloud + "-eo")
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
