package autossl

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type sdkResponse interface {
	GetCode() int
	GetStatus() string
	GetMessage() string
}

type sdkResponseBase struct {
	Code      *int    `json:"code,omitempty"`
	Status    *string `json:"status,omitempty"`
	Message   *string `json:"message,omitempty"`
	RequestID *string `json:"requestId,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

func (r *sdkResponseBase) GetStatus() string {
	if r.Status == nil {
		return ""
	}

	return *r.Status
}

func (r *sdkResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)

// 表示证书列表项的数据结构。
type CertificateInfo struct {
	CertId        string    `json:"certId"`
	Url           string    `json:"url"`
	CreateTime    string    `json:"createTime"`
	CertStartTime time.Time `json:"-"`
	CertEndTime   time.Time `json:"-"`
	CrtBrand      string    `json:"crtBrand"`
}

type certificateInfoRaw struct {
	CertId        string `json:"certId"`
	Url           string `json:"url"`
	CreateTime    any    `json:"createTime"`
	CertStartTime any    `json:"certStartTime"`
	CertEndTime   any    `json:"certEndTime"`
	CrtBrand      string `json:"crtBrand"`
}

func (r *certificateInfoRaw) toModel() *CertificateInfo {
	return &CertificateInfo{
		CertId:        r.CertId,
		Url:           r.Url,
		CreateTime:    toString(r.CreateTime),
		CertStartTime: parseCertTime(r.CertStartTime),
		CertEndTime:   parseCertTime(r.CertEndTime),
		CrtBrand:      r.CrtBrand,
	}
}

// 表示证书下载数据的数据结构。
type CertificateKeypair struct {
	CertPEM    string `json:"pem"`
	PrivkeyPEM string `json:"key"`
}

// 从响应负载中解析证书列表。
//
// 实测成功响应结构（PageHelper 分页格式）：
//
//	{"code":200,"data":{"pageNum":1,...,"total":N,"list":[{...}]},"message":"SUCCESS","status":"AUTO_SSL_SUCCESS"}
func parseCertificateItems(payload []byte) ([]*CertificateInfo, error) {
	var res struct {
		sdkResponseBase
		Data *struct {
			List []*certificateInfoRaw `json:"list"`
		} `json:"data,omitempty"`
	}
	if err := json.Unmarshal(payload, &res); err != nil {
		return nil, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	}

	if res.Data == nil {
		return nil, fmt.Errorf("sdkerr: no certificate list found in response")
	}

	items := make([]*CertificateInfo, 0, len(res.Data.List))
	for _, itemRaw := range res.Data.List {
		if itemRaw != nil {
			items = append(items, itemRaw.toModel())
		}
	}

	return items, nil
}

// 从响应负载中解析证书内容。
//
// 实测成功响应结构：
//
//	{"code":200,"data":{"pem":"-----BEGIN CERTIFICATE-----...","key":"-----BEGIN PRIVATE KEY-----..."}}
func parseCertificateKeypair(payload []byte) (*CertificateKeypair, error) {
	var res struct {
		sdkResponseBase
		Data *CertificateKeypair `json:"data,omitempty"`
	}
	if err := json.Unmarshal(payload, &res); err != nil {
		return nil, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	}

	if res.Data == nil {
		return nil, fmt.Errorf("sdkerr: no certificate keypair found in response")
	}

	return res.Data, nil
}

func toString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case float64:
		return strconv.FormatInt(int64(value), 10)
	default:
		return ""
	}
}

// 解析证书时间字段。
//
// 官方文档仅标注类型为 `Date`，未说明具体格式，此处按常见形态兼容解析：
// 毫秒级 Unix 时间戳、RFC3339、"2006-01-02 15:04:05"、"2006-01-02" 等。
// 解析失败时返回零值时间。
func parseCertTime(v any) time.Time {
	var certTime time.Time

	switch value := v.(type) {
	case float64:
		ms := int64(value)
		if ms > 1_000_000_000_000 {
			certTime = time.UnixMilli(ms)
		} else {
			certTime = time.Unix(ms, 0)
		}
	case string:
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"} {
			if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
				certTime = t
				break
			}
		}
	}

	return certTime
}
