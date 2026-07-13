package v1

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	vod "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1"
)

func GenReqDefForShowHttpsConfig() *def.HttpRequestDef {
	return vod.GenReqDefForShowHttpsConfig()
}

func GenReqDefForUpdateHttpsConfig() *def.HttpRequestDef {
	return vod.GenReqDefForUpdateHttpsConfig()
}
