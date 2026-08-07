package tencentcloud

import (
	"strings"
)

// 判断 API 接口端点是否属于腾讯云国际版。
func IsIntlAPIEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "intl.tencentcloudapi.com")
}

// 判断 EdgeOne 接口端点是否属于腾讯云国际版。
func IsIntlEdgeOneEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "edgeone.ai")
}
