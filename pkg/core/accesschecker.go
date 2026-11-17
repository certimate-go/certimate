package core

import (
	"context"
	"errors"
	"time"
)

type AccessChecker interface {
	ProviderType() string
	Scope() []string
	Check(ctx context.Context, config map[string]any) (*AccessCheckResult, error)
}

const (
	AccessCheckStatusPassed       = "passed"
	AccessCheckStatusFailed       = "failed"
	AccessCheckStatusInconclusive = "inconclusive"
	AccessCheckStatusUnavailable  = "unavailable"
)

const (
	AccessCheckItemStatusPassed  = "passed"
	AccessCheckItemStatusFailed  = "failed"
	AccessCheckItemStatusSkipped = "skipped"
)

const (
	AccessCheckReasonAuthRejected   = "auth_rejected"
	AccessCheckReasonConnectFailed  = "connect_failed"
	AccessCheckReasonTlsFailed      = "tls_failed"
	AccessCheckReasonTimeout        = "timeout"
	AccessCheckReasonCanceled       = "canceled"
	AccessCheckReasonProbeForbidden = "probe_forbidden"
	AccessCheckReasonRemoteError    = "remote_error"
)

const (
	AccessCheckItemConnect = "connect"
	AccessCheckItemAuth    = "auth"
)

const AccessCheckDefaultBudget = 20 * time.Second

var ErrAccessCheckConfigInvalid = errors.New("config_invalid")

type AccessCheckItem struct {
	Id     string
	Status string
}

type AccessCheckResult struct {
	Status      string
	ReasonCode  string
	Items       []AccessCheckItem
	Diagnostics []string
}
