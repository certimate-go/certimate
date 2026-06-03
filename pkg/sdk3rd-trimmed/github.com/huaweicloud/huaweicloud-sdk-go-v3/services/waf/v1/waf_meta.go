package v1

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	waf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
)

func GenReqDefForCreateCertificate() *def.HttpRequestDef {
	return waf.GenReqDefForCreateCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return waf.GenReqDefForListCertificates()
}

func GenReqDefForListHost() *def.HttpRequestDef {
	return waf.GenReqDefForListHost()
}

func GenReqDefForListPremiumHost() *def.HttpRequestDef {
	return waf.GenReqDefForListPremiumHost()
}

func GenReqDefForShowCertificate() *def.HttpRequestDef {
	return waf.GenReqDefForShowCertificate()
}

func GenReqDefForUpdateCertificate() *def.HttpRequestDef {
	return waf.GenReqDefForUpdateCertificate()
}

func GenReqDefForUpdateHost() *def.HttpRequestDef {
	return waf.GenReqDefForUpdateHost()
}

func GenReqDefForUpdatePremiumHost() *def.HttpRequestDef {
	return waf.GenReqDefForUpdatePremiumHost()
}
