package v3

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwelb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
)

func GenReqDefForCreateCertificate() *def.HttpRequestDef {
	return hwelb.GenReqDefForCreateCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return hwelb.GenReqDefForListCertificates()
}

func GenReqDefForListListeners() *def.HttpRequestDef {
	return hwelb.GenReqDefForListListeners()
}

func GenReqDefForShowCertificate() *def.HttpRequestDef {
	return hwelb.GenReqDefForShowCertificate()
}

func GenReqDefForShowListener() *def.HttpRequestDef {
	return hwelb.GenReqDefForShowListener()
}

func GenReqDefForShowLoadBalancer() *def.HttpRequestDef {
	return hwelb.GenReqDefForShowLoadBalancer()
}

func GenReqDefForUpdateCertificate() *def.HttpRequestDef {
	return hwelb.GenReqDefForUpdateCertificate()
}

func GenReqDefForUpdateListener() *def.HttpRequestDef {
	return hwelb.GenReqDefForUpdateListener()
}
