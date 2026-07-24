package plugin

import "testing"

func TestSemverCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.1.0", "0.1.0", 0},
		{"0.1.0", "0.2.0", -1},
		{"0.2.0", "0.1.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"v0.1.0", "0.1.0", 0},
		{"0.1", "0.1.0", 0},
		{"1.2.3-rc", "1.2.3", 0},
	}
	for _, c := range cases {
		got := semverCompare(c.a, c.b)
		sign := 0
		if got < 0 {
			sign = -1
		} else if got > 0 {
			sign = 1
		}
		if sign != c.want {
			t.Errorf("semverCompare(%q,%q)=%d want %d", c.a, c.b, sign, c.want)
		}
	}
}

func TestSemverLess(t *testing.T) {
	if !semverLess("0.1.0", "0.2.0") {
		t.Fatal("0.1.0 < 0.2.0 expected")
	}
	if semverLess("0.2.0", "0.1.0") {
		t.Fatal("0.2.0 < 0.1.0 not expected")
	}
}
