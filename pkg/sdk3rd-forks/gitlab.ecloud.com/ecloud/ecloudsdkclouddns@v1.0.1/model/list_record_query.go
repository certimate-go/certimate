// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type ListRecordQuery struct {
	position.Query
	// 页大小
	PageSize *int32 `json:"pageSize,omitempty"`

	// 当前页
	CurrentPage *int32 `json:"currentPage,omitempty"`
}
