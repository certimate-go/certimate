package filepath

import (
	stdfilepath "path/filepath"
	"strings"
)

// 与标准库中的 [filepath.Dir] 类似，但会尝试保留原有的路径分隔符。
//
// 入参:
//   - path: 文件路径。
//
// 出参:
//   - 目录路径。
func Dir(path string) string {
	const SEP_WIN = "\\"
	const SEP_UNIX = "/"

	sep := SEP_UNIX
	if strings.Contains(path, SEP_WIN) && !strings.Contains(path, SEP_UNIX) {
		sep = SEP_WIN
	}

	dir := stdfilepath.Dir(path)

	if sep != SEP_UNIX && strings.Contains(dir, SEP_UNIX) {
		dir = strings.ReplaceAll(dir, SEP_UNIX, sep)
	} else if sep != SEP_WIN && strings.Contains(dir, SEP_WIN) {
		dir = strings.ReplaceAll(dir, SEP_WIN, sep)
	}

	return dir
}
