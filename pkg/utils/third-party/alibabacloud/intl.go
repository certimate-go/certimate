package alibabacloud

import (
	"strings"
)

// 判断地域是否属于阿里云国际版。
func IsIntlRegion(region string) bool {
	region = strings.TrimSpace(region)
	return region != "" && !strings.HasPrefix(region, "cn-")
}
