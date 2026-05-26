package matrix

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const uiaPollInterval = 3 * time.Second

type loginAPIResponse struct {
	AccessToken string `json:"access_token"`
	DeviceID    string `json:"device_id"`
	UserID      string `json:"user_id"`
	Session     string `json:"session"`
	Flows       []struct {
		Stages []string `json:"stages"`
	} `json:"flows"`
	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

func (n *Notifier) loginWithSession(ctx context.Context, base string) (*LoginResult, error) {
	localpart, server, err := parseUserID(n.config.UserId, base)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(n.config.Password) == "" {
		return nil, fmt.Errorf("matrix: password is required for password auth")
	}

	deviceID := stableDeviceID(n.config)
	n.logger.Info(
		"matrix: logging in",
		slog.String("user", localpart),
		slog.String("server", server),
		slog.String("deviceId", deviceID),
	)

	body := map[string]any{
		"type": "m.login.password",
		"identifier": map[string]any{
			"type": "m.id.user",
			"user": localpart,
		},
		"password":                    n.config.Password,
		"initial_device_display_name": "Certimate Notifications",
		"device_id":                   deviceID,
	}

	client := n.httpClient.SetTimeout(matrixLoginTimeout)
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, matrixLoginTimeout)
		defer cancel()
		deadline, _ = ctx.Deadline()
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, status, raw, err := postLogin(ctx, client, base, body)
		if err != nil {
			return nil, err
		}

		if status >= 200 && status < 300 && resp.AccessToken != "" {
			outDevice := resp.DeviceID
			if outDevice == "" {
				outDevice = deviceID
			}
			n.logger.Info(
				"matrix: login ok",
				slog.String("userId", resp.UserID),
				slog.String("deviceId", outDevice),
			)
			return &LoginResult{
				AccessToken: resp.AccessToken,
				DeviceID:    outDevice,
				UserID:      resp.UserID,
			}, nil
		}

		if resp.Session != "" {
			body["session"] = resp.Session
			n.logger.Info(
				"matrix: waiting for interactive auth (e.g. MFA approval on Element)",
				slog.String("session", resp.Session),
				slog.Any("flows", resp.Flows),
			)
			if time.Until(deadline) < uiaPollInterval {
				return nil, &AuthError{
					Code:    "M_UIA_TIMEOUT",
					Message: "MFA / login approval timed out — confirm the login in Element and try again",
				}
			}
			if err := sleepWithContext(ctx, uiaPollInterval); err != nil {
				return nil, err
			}
			continue
		}

		return nil, formatLoginError(status, raw)
	}
}

func postLogin(ctx context.Context, client *resty.Client, base string, body map[string]any) (*loginAPIResponse, int, string, error) {
	var resp loginAPIResponse
	r, err := client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&resp).
		Post(base + "/_matrix/client/v3/login")
	if err != nil {
		return nil, 0, "", fmt.Errorf("matrix login request: %w", err)
	}
	raw := r.String()
	if resp.ErrCode == "" && resp.Error == "" && r.IsError() {
		_ = json.Unmarshal([]byte(raw), &resp)
	}
	return &resp, r.StatusCode(), raw, nil
}
