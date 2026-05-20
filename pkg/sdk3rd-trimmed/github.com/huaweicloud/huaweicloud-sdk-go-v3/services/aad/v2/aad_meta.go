package v2

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwaad "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2"
)

func GenReqDefForListInstanceDomains() *def.HttpRequestDef {
	return hwaad.GenReqDefForListInstanceDomains()
}
