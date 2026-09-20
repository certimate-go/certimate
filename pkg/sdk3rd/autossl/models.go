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
// 官方文档未给出成功响应的完整外层结构，此处按常见形态兼容解析：
//   - `data` 为数组；
//   - `data` 为对象，含 `list`、`records`、`rows`、`items`、`content` 等字段之一；
//   - 顶层直接为数组。
func parseCertificateItems(payload []byte) ([]*CertificateInfo, error) {
	var topAny any
	if err := json.Unmarshal(payload, &topAny); err != nil {
		return nil, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	}

	topMap, _ := topAny.(map[string]any)
	candidates := []any{topAny}
	if topMap != nil {
		if data, ok := topMap["data"]; ok {
			candidates = append(candidates, data)
			if dataMap, ok := data.(map[string]any); ok {
				for _, key := range []string{"list", "records", "rows", "items", "content", "certList"} {
					if v, ok := dataMap[key]; ok {
						candidates = append(candidates, v)
					}
				}
			}
		}
	}

	for _, candidate := range candidates {
		if arr, ok := candidate.([]any); ok {
			items := make([]*CertificateInfo, 0, len(arr))
			for _, elem := range arr {
				elemb, err := json.Marshal(elem)
				if err != nil {
					continue
				}
				var itemRaw certificateInfoRaw
				if err := json.Unmarshal(elemb, &itemRaw); err != nil {
					continue
				}
				items = append(items, itemRaw.toModel())
			}
			return items, nil
		}
	}

	return nil, fmt.Errorf("sdkerr: unrecognized certificate list structure in response")
}

// 从响应负载中解析证书内容。
//
// 官方文档未给出成功响应的完整外层结构，此处按常见形态兼容解析：
//   - `data` 为对象，含 `pem`、`key` 字段；
//   - 顶层直接为对象，含 `pem`、`key` 字段。
func parseCertificateKeypair(payload []byte) (*CertificateKeypair, error) {
	var topAny any
	if err := json.Unmarshal(payload, &topAny); err != nil {
		return nil, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	}

	topMap, _ := topAny.(map[string]any)
	if topMap == nil {
		return nil, fmt.Errorf("sdkerr: unrecognized certificate keypair structure in response")
	}

	candidates := []map[string]any{topMap}
	if data, ok := topMap["data"].(map[string]any); ok {
		candidates = append([]map[string]any{data}, candidates...)
	}

	for _, candidate := range candidates {
		pem, hasPem := candidate["pem"].(string)
		key, hasKey := candidate["key"].(string)
		if hasPem && hasKey {
			return &CertificateKeypair{CertPEM: pem, PrivkeyPEM: key}, nil
		}
	}

	return nil, fmt.Errorf("sdkerr: unrecognized certificate keypair structure in response")
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
