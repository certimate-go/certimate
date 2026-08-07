package engine

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/domain"
)

func TestShouldRenewByARI(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		windowStart time.Time
		windowEnd   time.Time
		want        bool
	}{
		{
			name:        "inactive window does not renew",
			windowStart: now.Add(time.Hour),
			windowEnd:   now.Add(2 * time.Hour),
			want:        false,
		},
		{
			name:        "active window renews",
			windowStart: now.Add(-time.Hour),
			windowEnd:   now.Add(time.Hour),
			want:        true,
		},
		{
			name:        "passed window renews",
			windowStart: now.Add(-2 * time.Hour),
			windowEnd:   now.Add(-time.Hour),
			want:        true,
		},
		{
			name: "zero window does not renew",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRenewByARI(tt.windowStart, tt.windowEnd, now); got != tt.want {
				t.Fatalf("shouldRenewByARI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateARI(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	ne := &bizApplyNodeExecutor{nodeExecutor: nodeExecutor{logger: slog.Default()}}

	t.Run("supported active window triggers renewal", func(t *testing.T) {
		fetch := func() (*certacme.ARIInfo, error) {
			return &certacme.ARIInfo{WindowStart: now.Add(-time.Hour), WindowEnd: now.Add(time.Hour), Supported: true}, nil
		}
		if !ne.evaluateARI(now, &domain.Certificate{}, fetch) {
			t.Fatal("expected ARI renewal within the active window")
		}
	})

	t.Run("unsupported CA does not trigger", func(t *testing.T) {
		fetch := func() (*certacme.ARIInfo, error) {
			return &certacme.ARIInfo{Supported: false}, nil
		}
		if ne.evaluateARI(now, &domain.Certificate{}, fetch) {
			t.Fatal("expected no ARI renewal when CA is unsupported")
		}
	})

	t.Run("fetch error does not trigger and does not block", func(t *testing.T) {
		fetch := func() (*certacme.ARIInfo, error) {
			return nil, errors.New("503 service unavailable")
		}
		if ne.evaluateARI(now, &domain.Certificate{}, fetch) {
			t.Fatal("expected no ARI renewal on fetch error")
		}
	})

	t.Run("queries every run regardless of past refresh", func(t *testing.T) {
		calls := 0
		fetch := func() (*certacme.ARIInfo, error) {
			calls++
			return &certacme.ARIInfo{WindowStart: now.Add(-time.Hour), WindowEnd: now.Add(time.Hour), Supported: true}, nil
		}
		_ = ne.evaluateARI(now, &domain.Certificate{}, fetch)
		_ = ne.evaluateARI(now, &domain.Certificate{}, fetch)
		if calls != 2 {
			t.Fatalf("expected renewal-info fetched on every run (2 calls), got %d", calls)
		}
	})
}
