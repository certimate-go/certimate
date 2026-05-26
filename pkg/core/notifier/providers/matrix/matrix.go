package matrix

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/pkg/core/notifier"
)

const (
	authModeToken    = "token"
	authModePassword = "password"
)

// matrixLoginTimeout is used for password login (slow homeservers).
const matrixLoginTimeout = 180 * time.Second

// verifyQuickRequestTimeout covers homeserver discovery, whoami, joined_rooms.
const verifyQuickRequestTimeout = 30 * time.Second

// verifyPauseAfterHomeserver separates discovery from login on the homeserver.
const verifyPauseAfterHomeserver = 2 * time.Second

type NotifierConfig struct {
	// Homeserver URL from Element (e.g. https://m.srvdev.ru).
	HomeserverUrl string `json:"homeserverUrl"`
	// "token" (bot or user access token) or "password".
	AuthMode string `json:"authMode,omitempty"`
	// Matrix access token (recommended for bots).
	AccessToken string `json:"accessToken,omitempty"`
	// SessionAccessToken is reused after password+MFA login (password auth mode).
	SessionAccessToken string `json:"sessionAccessToken,omitempty"`
	// SessionDeviceId stabilizes the Matrix device (one MFA approval).
	SessionDeviceId string `json:"sessionDeviceId,omitempty"`
	// MXID (@user:server) or localpart (user).
	UserId   string `json:"userId,omitempty"`
	Password string `json:"password,omitempty"`
	// Room ID (!room:server) — default target.
	RoomId string `json:"roomId,omitempty"`
}

type Notifier struct {
	config     *NotifierConfig
	logger     *slog.Logger
	httpClient *resty.Client
}

var _ notifier.Provider = (*Notifier)(nil)

func NewNotifier(config *NotifierConfig) (*Notifier, error) {
	if config == nil {
		return nil, errors.New("the configuration of the notifier provider is nil")
	}
	return &Notifier{
		config: config,
		logger: slog.Default(),
		httpClient: resty.New().
			SetHeader("Content-Type", "application/json").
			SetHeader("User-Agent", app.AppUserAgent),
	}, nil
}

func (n *Notifier) SetLogger(logger *slog.Logger) {
	if logger == nil {
		n.logger = slog.New(slog.DiscardHandler)
	} else {
		n.logger = logger
	}
}

func (n *Notifier) Notify(ctx context.Context, subject string, message string) (*notifier.NotifyResult, error) {
	roomID := strings.TrimSpace(n.config.RoomId)
	if roomID == "" {
		return nil, errors.New("matrix: room id is required")
	}

	base, err := resolveClientBaseURL(ctx, n.httpClient, n.config.HomeserverUrl)
	if err != nil {
		return nil, err
	}

	token, _, _, session, err := acquireCredentials(ctx, n.httpClient, n.config, base, n.logger)
	if err != nil {
		return nil, err
	}
	if session != nil {
		SetPendingSession(n.config, *session)
	}

	body := strings.TrimSpace(subject)
	if message != "" {
		if body != "" {
			body += "\n\n"
		}
		body += message
	}

	n.logger.Info(
		"matrix: sending message",
		slog.String("homeserver", base),
		slog.String("roomId", roomID),
	)

	if err := n.sendText(ctx, base, token, roomID, body); err != nil {
		return nil, err
	}

	n.logger.Info("matrix: message sent")
	return &notifier.NotifyResult{}, nil
}

func (n *Notifier) sendText(ctx context.Context, base, token, roomID, body string) error {
	txnID := newTxnID()
	path := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		base, url.PathEscape(roomID), url.PathEscape(txnID))

	payload := map[string]any{
		"msgtype": "m.text",
		"body":    body,
	}

	r, err := n.httpClient.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(payload).
		Put(path)
	if err != nil {
		return fmt.Errorf("matrix send message: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("matrix send message: status %d: %s", r.StatusCode(), r.String())
	}
	return nil
}

func resolveClientBaseURL(ctx context.Context, client *resty.Client, homeserver string) (string, error) {
	base, _, err := resolveClientBaseURLWithDetail(ctx, client, homeserver)
	if err != nil {
		return "", fmt.Errorf("matrix: %w", err)
	}
	return base, nil
}

func parseUserID(raw, homeserverBase string) (localpart, server string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("matrix: user id is required for password auth")
	}
	if strings.HasPrefix(raw, "@") {
		parts := strings.SplitN(raw[1:], ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("matrix: invalid mxid %q", raw)
		}
		return parts[0], parts[1], nil
	}
	server = serverFromHomeserverURL(homeserverBase)
	if server == "" {
		return "", "", errors.New("matrix: cannot derive server name from homeserver url")
	}
	return raw, server, nil
}

func serverFromHomeserverURL(base string) string {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" {
		return ""
	}
	host := u.Hostname()
	return host
}

func newTxnID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("certimate_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
