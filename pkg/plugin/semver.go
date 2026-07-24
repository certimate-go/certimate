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
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func semverParts(v string) [3]int64 {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	var out [3]int64
	for i, part := range strings.SplitN(v, ".", 3) {
		if i >= 3 {
			break
		}
		num, _, _ := strings.Cut(part, "-")
		n, _ := strconv.ParseInt(num, 10, 64)
		out[i] = n
	}
	return out
}
