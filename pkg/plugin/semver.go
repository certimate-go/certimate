package plugin

import (
	"strconv"
	"strings"
)

func semverLess(a, b string) bool {
	return semverCompare(a, b) < 0
}

func semverCompare(a, b string) int {
	pa := semverParts(a)
	pb := semverParts(b)
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var x, y int64
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func semverParts(v string) []int64 {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		num, _, _ := strings.Cut(part, "-")
		n, _ := strconv.ParseInt(num, 10, 64)
		out = append(out, n)
	}
	return out
}
