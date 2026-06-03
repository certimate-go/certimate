package v3

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	elb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
)

func GenReqDefForCreateCertificate() *def.HttpRequestDef {
	return elb.GenReqDefForCreateCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return elb.GenReqDefForListCertificates()
}

func GenReqDefForListListeners() *def.HttpRequestDef {
	return elb.GenReqDefForListListeners()
}

func GenReqDefForShowCertificate() *def.HttpRequestDef {
	return elb.GenReqDefForShowCertificate()
}

func GenReqDefForShowListener() *def.HttpRequestDef {
	return elb.GenReqDefForShowListener()
}

func GenReqDefForShowLoadBalancer() *def.HttpRequestDef {
	return elb.GenReqDefForShowLoadBalancer()
}

func GenReqDefForUpdateCertificate() *def.HttpRequestDef {
	return elb.GenReqDefForUpdateCertificate()
}

func GenReqDefForUpdateListener() *def.HttpRequestDef {
	return elb.GenReqDefForUpdateListener()
}
