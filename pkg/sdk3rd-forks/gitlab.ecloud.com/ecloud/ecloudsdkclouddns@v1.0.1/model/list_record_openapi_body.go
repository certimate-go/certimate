// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/position"
)

type ListRecordOpenapiBody struct {
	position.Body
	// 域名
	DomainName string `json:"domainName"`
}
