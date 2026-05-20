// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type DeleteRecordOpenapiBody struct {
	position.Body
	// 待删除的解析记录ID请求体
	RecordIdList []string `json:"recordIdList"`
}
