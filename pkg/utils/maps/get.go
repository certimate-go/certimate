package maps

import (
	"strconv"
	"strings"
)

// 以字符串形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是字符串，则返回空字符串。
func GetString(dict map[string]any, key string) string {
	return GetOrDefaultString(dict, key, "")
}

// 以字符串形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、值的类型不是字符串、或者值为零值，则返回默认值。
func GetOrDefaultString(dict map[string]any, key string, defaultValue string) string {
	if dict == nil {
		return defaultValue
	}

	if value, ok := dict[key]; ok {
		if result, ok := value.(string); ok {
			if result != "" {
				return result
			}
		}
	}

	return defaultValue
}

// 以字符串形式从字典中获取指定键的值，并按指定分隔符拆分。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是字符串，则返回空字符串切片。
func GetStringsBySplit(dict map[string]any, key string, sep string) []string {
	return GetOrDefaultStringsBySplit(dict, key, sep, make([]string, 0))
}

// 以字符串形式从字典中获取指定键的值，并按指定分隔符拆分。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、值的类型不是字符串、或者值为零值，则返回默认值。
func GetOrDefaultStringsBySplit(dict map[string]any, key string, sep string, defaultValue []string) []string {
	str := GetString(dict, key)
	if str != "" {
		return strings.Split(str, sep)
	}

	return defaultValue
}

// 以整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是整数，则返回 0。
func GetInt(dict map[string]any, key string) int {
	return GetOrDefaultInt(dict, key, 0)
}

// 以整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、值的类型不是整数、或者值为零值，则返回默认值。
func GetOrDefaultInt(dict map[string]any, key string, defaultValue int) int {
	if dict == nil {
		return defaultValue
	}

	if value, ok := dict[key]; ok {
		var result int

		switch v := value.(type) {
		case int:
			result = v
		case int8:
			result = int(v)
		case int16:
			result = int(v)
		case int32:
			result = int(v)
		case int64:
			result = int(v)
		case uint:
			result = int(v)
		case uint8:
			result = int(v)
		case uint16:
			result = int(v)
		case uint32:
			result = int(v)
		case uint64:
			result = int(v)
		case float32:
			result = int(v)
		case float64:
			result = int(v)
		case string:
			// 兼容字符串类型的值
			if t, err := strconv.ParseInt(v, 10, 64); err == nil {
				result = int(t)
			}
		}

		if result != 0 {
			return result
		}
	}

	return defaultValue
}

// 以 32 位整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是 32 位整数，则返回 0。
func GetInt32(dict map[string]any, key string) int32 {
	return GetOrDefaultInt32(dict, key, 0)
}

// 以 32 位整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、值的类型不是 32 位整数、或者值为零值，则返回默认值。
func GetOrDefaultInt32(dict map[string]any, key string, defaultValue int32) int32 {
	if dict == nil {
		return defaultValue
	}

	if value, ok := dict[key]; ok {
		var result int32

		switch v := value.(type) {
		case int:
			result = int32(v)
		case int8:
			result = int32(v)
		case int16:
			result = int32(v)
		case int32:
			result = v
		case int64:
			result = int32(v)
		case uint:
			result = int32(v)
		case uint8:
			result = int32(v)
		case uint16:
			result = int32(v)
		case uint32:
			result = int32(v)
		case uint64:
			result = int32(v)
		case float32:
			result = int32(v)
		case float64:
			result = int32(v)
		case string:
			// 兼容字符串类型的值
			if t, err := strconv.ParseInt(v, 10, 32); err == nil {
				result = int32(t)
			}
		}

		if result != 0 {
			return result
		}
	}

	return defaultValue
}

// 以 64 位整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是 64 位整数，则返回 0。
func GetInt64(dict map[string]any, key string) int64 {
	return GetOrDefaultInt64(dict, key, 0)
}

// 以 64 位整数形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、值的类型不是 64 位整数、或者值为零值，则返回默认值。
func GetOrDefaultInt64(dict map[string]any, key string, defaultValue int64) int64 {
	if dict == nil {
		return defaultValue
	}

	if value, ok := dict[key]; ok {
		var result int64

		switch v := value.(type) {
		case int:
			result = int64(v)
		case int8:
			result = int64(v)
		case int16:
			result = int64(v)
		case int32:
			result = int64(v)
		case int64:
			result = v
		case uint:
			result = int64(v)
		case uint8:
			result = int64(v)
		case uint16:
			result = int64(v)
		case uint32:
			result = int64(v)
		case uint64:
			result = int64(v)
		case float32:
			result = int64(v)
		case float64:
			result = int64(v)
		case string:
			// 兼容字符串类型的值
			if t, err := strconv.ParseInt(v, 10, 64); err == nil {
				result = t
			}
		}

		if result != 0 {
			return result
		}
	}

	return defaultValue
}

// 以布尔形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是布尔，则返回 false。
func GetBool(dict map[string]any, key string) bool {
	return GetOrDefaultBool(dict, key, false)
}

// 以布尔形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//   - defaultValue: 默认值。
//
// 出参：
//   - 字典中键对应的值。如果指定键不存在、或者值的类型不是布尔，则返回默认值。
func GetOrDefaultBool(dict map[string]any, key string, defaultValue bool) bool {
	if dict == nil {
		return defaultValue
	}

	if value, ok := dict[key]; ok {
		if result, ok := value.(bool); ok {
			return result
		}

		// 兼容字符串类型的值
		if str, ok := value.(string); ok {
			if result, err := strconv.ParseBool(str); err == nil {
				return result
			}
		}
	}

	return defaultValue
}

// 以 `map[string]V` 形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的 `map[string]V` 对象。
func GetKVMap[V any](dict map[string]any, key string) map[string]V {
	if dict == nil {
		return make(map[string]V)
	}

	if val, ok := dict[key]; ok {
		if result, ok := val.(map[string]V); ok {
			return result
		}
	}

	return make(map[string]V)
}

// 以 `map[string]any` 形式从字典中获取指定键的值。
//
// 入参：
//   - dict: 字典。
//   - key: 键。
//
// 出参：
//   - 字典中键对应的 `map[string]any` 对象。
func GetKVMapAny(dict map[string]any, key string) map[string]any {
	return GetKVMap[any](dict, key)
}
