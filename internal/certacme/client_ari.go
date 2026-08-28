package certacme

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/acme/api"

	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type ARIInfo struct {
	WindowStart   time.Time
	WindowEnd     time.Time
	NextRefreshAt time.Time
	Supported     bool
}

func (c *ACMEClient) GetARIInfo(ctx context.Context, certPEM string) (*ARIInfo, error) {
	if certPEM == "" {
		return nil, fmt.Errorf("the certificate PEM is empty")
	}

	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate PEM: %w", err)
	}

	renewalInfo, err := c.client.Certificate.GetRenewalInfo(ctx, certX509)
	if err != nil {
		if errors.Is(err, api.ErrNoARI) {
			return &ARIInfo{Supported: false}, nil
		}
		return nil, err
	}

	return &ARIInfo{
		WindowStart:   renewalInfo.SuggestedWindow.Start,
		WindowEnd:     renewalInfo.SuggestedWindow.End,
		NextRefreshAt: time.Now().Add(renewalInfo.RetryAfter),
		Supported:     true,
	}, nil
}
