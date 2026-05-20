package v1

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwwaf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
)

func GenReqDefForCreateCertificate() *def.HttpRequestDef {
	return hwwaf.GenReqDefForCreateCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return hwwaf.GenReqDefForListCertificates()
}

func GenReqDefForListHost() *def.HttpRequestDef {
	return hwwaf.GenReqDefForListHost()
}

func GenReqDefForListPremiumHost() *def.HttpRequestDef {
	return hwwaf.GenReqDefForListPremiumHost()
}

func GenReqDefForShowCertificate() *def.HttpRequestDef {
	return hwwaf.GenReqDefForShowCertificate()
}

func GenReqDefForUpdateCertificate() *def.HttpRequestDef {
	return hwwaf.GenReqDefForUpdateCertificate()
}

func GenReqDefForUpdateHost() *def.HttpRequestDef {
	return hwwaf.GenReqDefForUpdateHost()
}

func GenReqDefForUpdatePremiumHost() *def.HttpRequestDef {
	return hwwaf.GenReqDefForUpdatePremiumHost()
}
