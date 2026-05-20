// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordOpenapiResponseBody struct {
	// 当前页的具体数据列表
	Data *[]ListRecordOpenapiResponseData `json:"data,omitempty"`

	// 总数据量
	TotalNum *int32 `json:"totalNum,omitempty"`

	// 总页数
	TotalPages *int32 `json:"totalPages,omitempty"`

	// 页大小
	PageSize *int32 `json:"pageSize,omitempty"`

	// 当前页码，从0开始，0表示第一页
	Page *int32 `json:"page,omitempty"`
}
