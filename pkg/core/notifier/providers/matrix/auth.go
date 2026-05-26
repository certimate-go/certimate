package matrix

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/go-resty/resty/v2"
)

func acquireCredentials(ctx context.Context, client *resty.Client, cfg *NotifierConfig, base string, logger *slog.Logger) (token, userID, detail string, session *SessionCredentials, err error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.AuthMode))
	if mode == "" {
		if strings.TrimSpace(cfg.AccessToken) != "" && strings.TrimSpace(cfg.SessionAccessToken) == "" {
			mode = authModeToken
		} else if strings.TrimSpace(cfg.sessionToken()) != "" {
			mode = authModePassword
		} else if strings.TrimSpace(cfg.Password) != "" {
			mode = authModePassword
		} else {
			mode = authModeToken
		}
	}

	switch mode {
	case authModeToken:
		token = strings.TrimSpace(cfg.AccessToken)
		if token == "" {
			return "", "", "", nil, fmt.Errorf("access token is required")
		}
		detail = "using configured access token"
	case authModePassword:
		n := &Notifier{config: cfg, httpClient: client, logger: logger}
		if logger == nil {
			n.logger = slog.Default()
		}

		token, session, err = resolveAccessToken(ctx, client, cfg, base)
		if err != nil {
			n.logger.Info("matrix: saved session invalid, logging in again", slog.String("error", err.Error()))
			token = ""
			session = nil
		}
		if token != "" {
			detail = "using saved session token"
			break
		}

		loginClient := resty.New().
			SetHeader("Content-Type", "application/json").
			SetHeader("User-Agent", app.AppUserAgent).
			SetTimeout(matrixLoginTimeout)
		n.httpClient = loginClient

		res, loginErr := n.loginWithSession(ctx, base)
		if loginErr != nil {
			return "", "", "", nil, loginErr
		}
		token = res.AccessToken
		session = &SessionCredentials{AccessToken: res.AccessToken, DeviceID: res.DeviceID}
		detail = "password login succeeded; session can be saved"
		SetPendingSession(cfg, *session)
	default:
		return "", "", "", nil, fmt.Errorf("unknown auth mode %q", mode)
	}

	userID, whoDetail, err := whoamiUserID(ctx, client, base, token)
	if err != nil {
		return "", "", detail, session, err
	}
	if whoDetail != "" {
		detail += "; " + whoDetail
	}
	return token, userID, detail, session, nil
}

func whoamiUserID(ctx context.Context, client *resty.Client, base, token string) (userID, detail string, err error) {
	var whoami struct {
		UserID string `json:"user_id"`
		Error  string `json:"error"`
	}
	r, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&whoami).
		Get(base + "/_matrix/client/v3/account/whoami")
	if err != nil {
		return "", "", fmt.Errorf("whoami request: %w", err)
	}
	if r.IsError() || whoami.UserID == "" {
		msg := whoami.Error
		if msg == "" {
			msg = r.String()
		}
		return "", "", fmt.Errorf("token invalid (%d): %s", r.StatusCode(), msg)
	}
	return whoami.UserID, "user " + whoami.UserID, nil
}
