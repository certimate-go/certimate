package maps

// 将源字典中的所有键值对复制到目标字典中。
//
// 入参：
//   - src: 源字典。
//   - dist: 目标字典。
func CopyTo[Map ~map[K]V, K comparable, V any](src Map, dist Map) {
	for k, v := range src {
		dist[k] = v
	}
}
