package qiniu

import (
	"fmt"
	"strings"
)

const qiniuHost = "https://api.qiniu.com"

func urlf(pathf string, pathargs ...any) string {
	path := fmt.Sprintf(pathf, pathargs...)
	path = strings.TrimPrefix(path, "/")
	return qiniuHost + "/" + path
}
