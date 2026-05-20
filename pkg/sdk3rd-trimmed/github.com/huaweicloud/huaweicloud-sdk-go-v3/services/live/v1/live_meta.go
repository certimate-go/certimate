package v1

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwlive "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1"
)

func GenReqDefForShowDomain() *def.HttpRequestDef {
	return hwlive.GenReqDefForShowDomain()
}

func GenReqDefForUpdateDomainHttpsCert() *def.HttpRequestDef {
	return hwlive.GenReqDefForUpdateDomainHttpsCert()
}
