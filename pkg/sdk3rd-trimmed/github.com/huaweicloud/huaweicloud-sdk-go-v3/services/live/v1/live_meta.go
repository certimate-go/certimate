package v1

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	live "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1"
)

func GenReqDefForShowDomain() *def.HttpRequestDef {
	return live.GenReqDefForShowDomain()
}

func GenReqDefForUpdateDomainHttpsCert() *def.HttpRequestDef {
	return live.GenReqDefForUpdateDomainHttpsCert()
}
