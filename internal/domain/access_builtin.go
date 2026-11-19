package domain

import "slices"

var builtinAccessProviderTypes = map[AccessProviderType]bool{
	AccessProviderTypeACMECA: true,

	AccessProviderTypeActalisSSL: true,

	AccessProviderTypeAliyun: true,

	AccessProviderTypeAWS:        true,
	AccessProviderTypeAzure:      true,
	AccessProviderTypeBaiduCloud: true,

	AccessProviderTypeBunny:      true,
	AccessProviderTypeBytePlus:   true,
	AccessProviderTypeCloudflare: true,

	AccessProviderTypeCMCCCloud: true,

	AccessProviderTypeCPanel:    true,
	AccessProviderTypeCTCCCloud: true,
	AccessProviderTypeCUCCCloud: true,

	AccessProviderTypeDigiCert:     true,
	AccessProviderTypeDigitalOcean: true,

	AccessProviderTypeEmail:  true,
	AccessProviderTypeFastly: true,
	AccessProviderTypeFTP:    true,

	AccessProviderTypeGcore:           true,
	AccessProviderTypeGlobalSignAtlas: true,

	AccessProviderTypeGoogleCloud:         true,
	AccessProviderTypeGoogleTrustServices: true,

	AccessProviderTypeHuaweiCloud: true,

	AccessProviderTypeJDCloud:            true,
	AccessProviderTypeLetsEncrypt:        true,
	AccessProviderTypeLetsEncryptStaging: true,
	AccessProviderTypeLinode:             true,
	AccessProviderTypeLiteSSL:            true,
	AccessProviderTypeLocal:              true,

	AccessProviderTypeNetlify: true,

	AccessProviderTypeOracleCloud: true,

	AccessProviderTypeQingCloud: true,
	AccessProviderTypeRainYun:   true,

	AccessProviderTypeS3:      true,
	AccessProviderTypeSectigo: true,

	AccessProviderTypeSSH:    true,
	AccessProviderTypeSSLCOM: true,

	AccessProviderTypeTencentCloud: true,

	AccessProviderTypeUCloud:     true,
	AccessProviderTypeVercel:     true,
	AccessProviderTypeVolcEngine: true,

	AccessProviderTypeWebhook: true,

	AccessProviderTypeYandexCloud: true,
	AccessProviderTypeZeroSSL:     true,
}

func IsBuiltInAccessProviderType(t AccessProviderType) bool {
	return builtinAccessProviderTypes[t]
}

func AllBuiltInAccessProviderTypes() []AccessProviderType {
	providerTypes := make([]AccessProviderType, 0, len(builtinAccessProviderTypes))
	for providerType := range builtinAccessProviderTypes {
		providerTypes = append(providerTypes, providerType)
	}
	slices.Sort(providerTypes)
	return providerTypes
}
