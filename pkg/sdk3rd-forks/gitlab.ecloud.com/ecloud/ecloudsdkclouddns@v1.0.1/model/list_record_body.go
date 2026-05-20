// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type ListRecordBody struct {
	position.Body
	// 域名
	DomainName string `json:"domainName"`

	// 可以匹配主机头rr、记录值value、备注description，并且是模糊搜索
	DataLike string `json:"dataLike,omitempty"`
}
