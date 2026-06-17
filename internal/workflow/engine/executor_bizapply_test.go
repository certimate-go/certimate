package engine

import (
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

func TestApplyARIInfoToCertificate(t *testing.T) {
	windowStart := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(time.Hour)
	nextRefreshAt := windowStart.Add(30 * time.Minute)

	certificate := &domain.Certificate{}
	applyARIInfoToCertificate(certificate, &certacme.ARIInfo{
		WindowStart:   windowStart,
		WindowEnd:     windowEnd,
		NextRefreshAt: nextRefreshAt,
		Supported:     true,
	})

	if !certificate.ARISupported {
		t.Fatal("ARISupported = false, want true")
	}
	if !certificate.ARIWindowStart.Equal(windowStart) || !certificate.ARIWindowEnd.Equal(windowEnd) || !certificate.ARINextRefreshAt.Equal(nextRefreshAt) {
		t.Fatal("ARI timestamps were not applied")
	}

	applyARIInfoToCertificate(certificate, &certacme.ARIInfo{Supported: false})
	if certificate.ARISupported {
		t.Fatal("ARISupported = true, want false")
	}
	if !certificate.ARIWindowStart.IsZero() || !certificate.ARIWindowEnd.IsZero() || !certificate.ARINextRefreshAt.IsZero() {
		t.Fatal("ARI timestamps were not cleared for unsupported CA")
	}
}

func TestCopyARIFields(t *testing.T) {
	source := &domain.Certificate{
		ARIWindowStart:   time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC),
		ARIWindowEnd:     time.Date(2026, 6, 9, 13, 0, 0, 0, time.UTC),
		ARINextRefreshAt: time.Date(2026, 6, 9, 12, 30, 0, 0, time.UTC),
		ARISupported:     true,
	}
	target := &domain.Certificate{}

	copyARIFields(target, source)
	if !target.ARISupported {
		t.Fatal("ARISupported = false, want true")
	}
	if !target.ARIWindowStart.Equal(source.ARIWindowStart) || !target.ARIWindowEnd.Equal(source.ARIWindowEnd) || !target.ARINextRefreshAt.Equal(source.ARINextRefreshAt) {
		t.Fatal("ARI timestamps were not copied")
	}
}
