// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type DeleteRecordBody struct {
	position.Body
	// 解析记录ID列表
	RecordIdList []string `json:"recordIdList"`
}
