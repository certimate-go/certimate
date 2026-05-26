package matrix

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
)

// SessionCredentials is a Matrix device session reused across notifications.
type SessionCredentials struct {
	AccessToken string
	DeviceID    string
}

// LoginResult is returned by password login (new or existing session).
type LoginResult struct {
	AccessToken string
	DeviceID    string
	UserID      string
}

var pendingSessions sync.Map

func sessionStoreKey(cfg *NotifierConfig) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(cfg.HomeserverUrl)) + "|" + strings.ToLower(strings.TrimSpace(cfg.UserId))))
	return hex.EncodeToString(h[:8])
}

// SetPendingSession stores credentials to persist into access config after notify.
func SetPendingSession(cfg *NotifierConfig, creds SessionCredentials) {
	if cfg == nil || creds.AccessToken == "" {
		return
	}
	pendingSessions.Store(sessionStoreKey(cfg), creds)
}

// TakePendingSession returns and clears credentials waiting to be saved.
func TakePendingSession(cfg *NotifierConfig) (SessionCredentials, bool) {
	if cfg == nil {
		return SessionCredentials{}, false
	}
	v, ok := pendingSessions.LoadAndDelete(sessionStoreKey(cfg))
	if !ok {
		return SessionCredentials{}, false
	}
	creds, ok := v.(SessionCredentials)
	return creds, ok
}

func resolveAccessToken(ctx context.Context, client *resty.Client, cfg *NotifierConfig, base string) (token string, creds *SessionCredentials, err error) {
	token = strings.TrimSpace(cfg.sessionToken())
	if token == "" {
		return "", nil, nil
	}

	if err := validateToken(ctx, client, base, token); err != nil {
		return "", nil, err
	}

	deviceID := strings.TrimSpace(cfg.SessionDeviceId)
	return token, &SessionCredentials{AccessToken: token, DeviceID: deviceID}, nil
}

func (cfg *NotifierConfig) sessionToken() string {
	if t := strings.TrimSpace(cfg.SessionAccessToken); t != "" {
		return t
	}
	mode := strings.ToLower(strings.TrimSpace(cfg.AuthMode))
	if mode == authModeToken || mode == "" && strings.TrimSpace(cfg.AccessToken) != "" {
		return strings.TrimSpace(cfg.AccessToken)
	}
	return ""
}

func stableDeviceID(cfg *NotifierConfig) string {
	if id := strings.TrimSpace(cfg.SessionDeviceId); id != "" {
		return id
	}
	h := sha256.Sum256([]byte("certimate|" + strings.ToLower(strings.TrimSpace(cfg.HomeserverUrl)) + "|" + strings.ToLower(strings.TrimSpace(cfg.UserId))))
	return "CERTIMATE_" + hex.EncodeToString(h[:6])
}

func validateToken(ctx context.Context, client *resty.Client, base, token string) error {
	var whoami struct {
		UserID string `json:"user_id"`
	}
	r, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&whoami).
		Get(base + "/_matrix/client/v3/account/whoami")
	if err != nil {
		return fmt.Errorf("whoami: %w", err)
	}
	if r.IsError() || whoami.UserID == "" {
		return fmt.Errorf("session token invalid (%d)", r.StatusCode())
	}
	return nil
}
