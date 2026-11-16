package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParsePurgeExpiredDays(t *testing.T) {
	cases := []struct {
		name    string
		raw     any
		want    int64
		wantErr bool
	}{
		{"missing", nil, 0, false},
		{"empty string", "", 0, false},
		{"zero float", float64(0), 0, false},
		{"positive float", float64(7), 7, false},
		{"int", 30, 30, false},
		{"int64", int64(365), 365, false},
		{"json number", json.Number("14"), 14, false},
		{"decimal string", "21", 21, false},
		{"padded string", " 7 ", 7, true},
		{"negative float", float64(-1), 0, true},
		{"negative string", "-7", 0, true},
		{"fractional float", float64(1.5), 0, true},
		{"fractional string", "1.5", 0, true},
		{"non numeric string", "abc", 0, true},
		{"hex string", "0x10", 0, true},
		{"plus sign string", "+7", 7, false},
		{"float infinity", json.Number("1e999"), 0, true},
		{"bool", true, 0, true},
		{"slice", []any{1}, 0, true},
		{"max safe", float64(106751), 106751, false},
		{"over safe range", float64(106752), 0, true},
		{"huge float", 1e18, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePurgeExpiredDays(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %v, got %d", tc.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %v: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParsePurgeExpiryDuration(t *testing.T) {
	d, err := ParsePurgeExpiryDuration(float64(7))
	if err != nil {
		t.Fatal(err)
	}
	if d != 7*24*time.Hour {
		t.Fatalf("got %s", d)
	}
	if _, err := ParsePurgeExpiryDuration("not-a-number"); err == nil {
		t.Fatal("expected error")
	}
}
