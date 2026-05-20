// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package model

type ListRecordResponseBody struct {
	// 总页数
	TotalPages *int32 `json:"totalPages,omitempty"`

	// 当前页码，从0开始，0表示第一页
	CurrentPage *int32 `json:"currentPage,omitempty"`

	// 当前页的具体数据列表
	Results *[]ListRecordResponseResults `json:"results,omitempty"`

	// 总数据量
	TotalElements *int64 `json:"totalElements,omitempty"`
}
