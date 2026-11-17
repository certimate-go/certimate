package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

const maxPurgeExpiredDays = int64(math.MaxInt64 / int64(24*time.Hour))

func ParsePurgeExpiredDays(raw any) (int64, error) {
	switch v := raw.(type) {
	case nil:
		return 0, nil
	case float64:
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("expiredDays must be an integer, got %v", v)
		}
		return parsePurgeExpiredDaysInt(int64(v))
	case int:
		return parsePurgeExpiredDaysInt(int64(v))
	case int64:
		return parsePurgeExpiredDaysInt(v)
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("expiredDays must be an integer, got %q", v.String())
		}
		return parsePurgeExpiredDaysInt(n)
	case string:
		s := v
		if s == "" {
			return 0, nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("expiredDays must be a decimal integer, got %q", s)
		}
		return parsePurgeExpiredDaysInt(n)
	default:
		return 0, fmt.Errorf("expiredDays must be a number or decimal integer string, got %T", raw)
	}
}

func parsePurgeExpiredDaysInt(n int64) (int64, error) {
	if n < 0 {
		return 0, fmt.Errorf("expiredDays must not be negative, got %d", n)
	}
	if n > maxPurgeExpiredDays {
		return 0, fmt.Errorf("expiredDays %d exceeds the safe maximum %d", n, maxPurgeExpiredDays)
	}
	return n, nil
}

func ParsePurgeExpiryDuration(raw any) (time.Duration, error) {
	days, err := ParsePurgeExpiredDays(raw)
	if err != nil {
		return 0, err
	}
	return time.Duration(days) * 24 * time.Hour, nil
}
