package maps

import (
	"strings"
)

// 在字典中递归地替换指定的字符串的值。
// 注意该函数会修改原始的字典。
//
// 入参：
//   - dict: 字典。
//   - oldStr：需要替换的字符串。
//   - newStr：替换后的字符串。
//
// 出参：
//   - dict: 替换后的字典。
func DeepReplaceValue(dict map[string]any, oldStr, newStr string) map[string]any {
	return deepReplaceMapValue(dict, oldStr, newStr).(map[string]any)
}

// 与 [DeepReplaceValue] 类似，但入参类型为 `any`。
func DeepReplaceValueUnsafe(dict any, oldStr, newStr string) any {
	return deepReplaceMapValue(dict, oldStr, newStr)
}

func deepReplaceMapValue(data any, oldStr, newStr string) any {
	switch v := data.(type) {
	case map[string]any:
		for k, va := range v {
			v[k] = deepReplaceMapValue(va, oldStr, newStr)
		}
	case []any:
		for i, va := range v {
			v[i] = deepReplaceMapValue(va, oldStr, newStr)
		}
	case []string:
		for i, vs := range v {
			v[i] = deepReplaceMapValue(vs, oldStr, newStr).(string)
		}
	case string:
		return strings.ReplaceAll(v, oldStr, newStr)
	}
	return data
}
