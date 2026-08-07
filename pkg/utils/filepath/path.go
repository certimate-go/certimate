package filepath

import (
	stdfilepath "path/filepath"
	"strings"
)

const (
	separatorOnWindows = "\\"
	separatorOnUnix    = "/"
)

// 与标准库中的 [filepath.Dir] 类似，但会尝试保留原有的路径分隔符。
//
// 入参:
//   - path: 文件路径。
//
// 出参:
//   - 目录路径。
func Dir(path string) string {
	sep := string(stdfilepath.Separator)
	if strings.Contains(path, separatorOnWindows) && !strings.Contains(path, separatorOnUnix) {
		sep = separatorOnWindows
	} else if strings.Contains(path, separatorOnUnix) && !strings.Contains(path, separatorOnWindows) {
		sep = separatorOnUnix
	}

	dir := stdfilepath.Dir(path)
	return normalizePath(sep, dir)
}

// 与标准库中的 [filepath.Join] 类似，但会尝试保留原有的路径分隔符。
//
// 入参:
//   - elem: 路径元素。
//
// 出参:
//   - 连接后的路径。
func Join(elem ...string) string {
	sep := string(stdfilepath.Separator)
	for _, e := range elem {
		if strings.Contains(e, separatorOnWindows) && !strings.Contains(e, separatorOnUnix) {
			sep = separatorOnWindows
			break
		} else if strings.Contains(e, separatorOnUnix) && !strings.Contains(e, separatorOnWindows) {
			sep = separatorOnUnix
			break
		}
	}

	path := stdfilepath.Join(elem...)
	return normalizePath(sep, path)
}

func normalizePath(separator, path string) string {
	if separator != separatorOnUnix && strings.Contains(path, separatorOnUnix) {
		path = strings.ReplaceAll(path, separatorOnUnix, separator)
	} else if separator != separatorOnWindows && strings.Contains(path, separatorOnWindows) {
		path = strings.ReplaceAll(path, separatorOnWindows, separator)
	}

	return path
}
