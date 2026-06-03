package v2

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	cdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
)

func GenReqDefForListDomains() *def.HttpRequestDef {
	return cdn.GenReqDefForListDomains()
}

func GenReqDefForUpdateDomainMultiCertificates() *def.HttpRequestDef {
	return cdn.GenReqDefForUpdateDomainMultiCertificates()
}
