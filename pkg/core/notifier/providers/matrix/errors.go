package matrix

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AuthError is a Matrix login/auth failure with a machine-readable code.
type AuthError struct {
	Code          string
	RetryAfterSec int
	Message       string
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

type matrixAPIErrorBody struct {
	ErrCode      string `json:"errcode"`
	Error        string `json:"error"`
	RetryAfterMs int    `json:"retry_after_ms"`
}

func parseMatrixAPIError(statusCode int, rawBody string) error {
	rawBody = strings.TrimSpace(rawBody)
	var body matrixAPIErrorBody
	if rawBody != "" {
		_ = json.Unmarshal([]byte(rawBody), &body)
	}

	if body.ErrCode == "" && body.Error == "" {
		if rawBody == "" {
			return fmt.Errorf("matrix request failed (%d)", statusCode)
		}
		return fmt.Errorf("matrix request failed (%d): %s", statusCode, rawBody)
	}

	retrySec := 0
	if body.RetryAfterMs > 0 {
		retrySec = (body.RetryAfterMs + 999) / 1000
	}

	switch body.ErrCode {
	case "M_LIMIT_EXCEEDED":
		msg := "Too many login attempts — wait before trying again"
		if retrySec > 0 {
			msg = fmt.Sprintf("Too many login attempts — wait about %d s and try again", retrySec)
		}
		return &AuthError{Code: body.ErrCode, RetryAfterSec: retrySec, Message: msg}
	case "M_FORBIDDEN":
		return &AuthError{
			Code:    body.ErrCode,
			Message: "Invalid username or password",
		}
	case "M_USER_DEACTIVATED":
		return &AuthError{
			Code:    body.ErrCode,
			Message: "Account is deactivated",
		}
	case "M_UIA_TIMEOUT":
		return &AuthError{
			Code:    body.ErrCode,
			Message: body.Error,
		}
	default:
		msg := body.Error
		if msg == "" {
			msg = rawBody
		}
		return &AuthError{
			Code:    body.ErrCode,
			Message: fmt.Sprintf("%s (%d)", msg, statusCode),
		}
	}
}

func formatLoginError(statusCode int, rawBody string) error {
	if statusCode == 429 {
		return parseMatrixAPIError(statusCode, rawBody)
	}
	var body matrixAPIErrorBody
	if json.Unmarshal([]byte(strings.TrimSpace(rawBody)), &body) == nil && body.ErrCode != "" {
		return parseMatrixAPIError(statusCode, rawBody)
	}
	return fmt.Errorf("matrix login failed (%d): %s", statusCode, rawBody)
}
