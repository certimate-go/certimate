package v2

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwcdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
)

func GenReqDefForListDomains() *def.HttpRequestDef {
	return hwcdn.GenReqDefForListDomains()
}

func GenReqDefForUpdateDomainMultiCertificates() *def.HttpRequestDef {
	return hwcdn.GenReqDefForUpdateDomainMultiCertificates()
}
