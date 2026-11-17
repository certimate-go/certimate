package domain

var purgeProviderAliases = map[PurgeProviderType]PurgeProviderType{
	PurgeProviderTypeAliyunCAS:            PurgeProviderTypeAliyunCAS + "-purger",
	PurgeProviderTypeAliyunCLB:            PurgeProviderTypeAliyunCLB + "-purger",
	PurgeProviderTypeAliyunESA:            PurgeProviderTypeAliyunESA + "-purger",
	PurgeProviderTypeAWSACM:               PurgeProviderTypeAWSACM + "-purger",
	PurgeProviderTypeAWSIAM:               PurgeProviderTypeAWSIAM + "-purger",
	PurgeProviderTypeAzureKeyVault:        PurgeProviderTypeAzureKeyVault + "-purger",
	PurgeProviderTypeBaiduCloudCert:       PurgeProviderTypeBaiduCloudCert + "-purger",
	PurgeProviderTypeHuaweiCloudELB:       PurgeProviderTypeHuaweiCloudELB + "-purger",
	PurgeProviderTypeHuaweiCloudSCM:       PurgeProviderTypeHuaweiCloudSCM + "-purger",
	PurgeProviderTypeHuaweiCloudWAF:       PurgeProviderTypeHuaweiCloudWAF + "-purger",
	PurgeProviderTypeJDCloudSSL:           PurgeProviderTypeJDCloudSSL + "-purger",
	PurgeProviderTypeTencentCloudGAAP:     PurgeProviderTypeTencentCloudGAAP + "-purger",
	PurgeProviderTypeTencentCloudSSL:      PurgeProviderTypeTencentCloudSSL + "-purger",
	PurgeProviderTypeTencentCloudTSE:      PurgeProviderTypeTencentCloudTSE + "-purger",
	PurgeProviderTypeUCloudUSSL:           PurgeProviderTypeUCloudUSSL + "-purger",
	PurgeProviderTypeVolcEngineCertCenter: PurgeProviderTypeVolcEngineCertCenter + "-purger",
}

func ResolvePurgeProviderAlias(provider PurgeProviderType) PurgeProviderType {
	if resolved, ok := purgeProviderAliases[provider]; ok {
		return resolved
	}
	return provider
}
