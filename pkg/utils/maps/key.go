package maps

// 获取字典中的所有键组成的切片。
//
// 入参：
//   - dict: 字典。
//
// 出参：
//   - keys: 字典中的所有键组成的切片。
func Keys[Map ~map[K]V, K comparable, V any](dict Map) []K {
	keys := make([]K, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	return keys
}
