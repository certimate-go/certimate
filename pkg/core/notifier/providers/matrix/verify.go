package matrix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/go-resty/resty/v2"
)

// VerifyStep is one check in the connection test wizard.
type VerifyStep struct {
	Name          string `json:"name"`
	Ok            bool   `json:"ok"`
	Message       string `json:"message"`
	Detail        string `json:"detail,omitempty"`
	Code          string `json:"code,omitempty"`
	RetryAfterSec int    `json:"retryAfterSec,omitempty"`
}

// VerifyResult is returned by VerifyConnection.
type VerifyResult struct {
	Ok                 bool         `json:"ok"`
	UserId             string       `json:"userId,omitempty"`
	SessionAccessToken string       `json:"sessionAccessToken,omitempty"`
	SessionDeviceId    string       `json:"sessionDeviceId,omitempty"`
	Steps              []VerifyStep `json:"steps"`
}

// VerifyConnection checks homeserver, auth, and optionally room membership.
func VerifyConnection(ctx context.Context, cfg *NotifierConfig) (*VerifyResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("matrix: config is nil")
	}
	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetTimeout(verifyQuickRequestTimeout)

	res := &VerifyResult{Steps: make([]VerifyStep, 0, 4)}

	addStep := func(step VerifyStep) {
		res.Steps = append(res.Steps, step)
		if !step.Ok {
			res.Ok = false
		}
	}
	add := func(name, okMsg, failMsg, detail string, ok bool) {
		step := VerifyStep{Name: name, Ok: ok, Detail: detail}
		if ok {
			step.Message = okMsg
		} else {
			step.Message = failMsg
		}
		addStep(step)
	}
	addAuthErr := func(err error, detail string) {
		step := VerifyStep{Name: "auth", Ok: false, Detail: detail, Message: err.Error()}
		var authErr *AuthError
		if errors.As(err, &authErr) {
			step.Code = authErr.Code
			step.RetryAfterSec = authErr.RetryAfterSec
			step.Message = authErr.Message
		}
		addStep(step)
	}

	base, discoverDetail, err := resolveClientBaseURLWithDetail(ctx, client, cfg.HomeserverUrl)
	if err != nil {
		add("homeserver", "", err.Error(), discoverDetail, false)
		return res, nil
	}
	add("homeserver", "Homeserver reachable", "Homeserver unreachable", discoverDetail, true)

	if err := sleepWithContext(ctx, verifyPauseAfterHomeserver); err != nil {
		return res, err
	}

	token, userID, authDetail, session, err := acquireCredentials(ctx, client, cfg, base, slog.Default())
	if err != nil {
		addAuthErr(err, authDetail)
		return res, nil
	}
	res.UserId = userID
	if session != nil {
		res.SessionAccessToken = session.AccessToken
		res.SessionDeviceId = session.DeviceID
		SetPendingSession(cfg, *session)
	}
	add("auth", "Authentication successful", "Authentication failed", authDetail, true)

	roomID := strings.TrimSpace(cfg.RoomId)
	if roomID == "" {
		add("room", "Room ID not set (optional before save)", "", "", true)
		res.Ok = true
		for _, s := range res.Steps {
			if !s.Ok {
				res.Ok = false
				break
			}
		}
		return res, nil
	}

	roomOK, roomMsg, roomDetail, err := checkJoinedRoom(ctx, client, base, token, roomID)
	if err != nil {
		add("room", "", err.Error(), roomDetail, false)
		return res, nil
	}
	if roomOK {
		add("room", roomMsg, "", roomDetail, true)
	} else {
		add("room", "", roomMsg, roomDetail, false)
	}

	res.Ok = true
	for _, s := range res.Steps {
		if !s.Ok {
			res.Ok = false
			break
		}
	}
	return res, nil
}

func resolveClientBaseURLWithDetail(ctx context.Context, client *resty.Client, homeserver string) (base, detail string, err error) {
	homeserver = strings.TrimSpace(homeserver)
	if homeserver == "" {
		return "", "", fmt.Errorf("homeserver url is required")
	}
	if !strings.HasPrefix(homeserver, "http://") && !strings.HasPrefix(homeserver, "https://") {
		homeserver = "https://" + homeserver
	}
	entered := strings.TrimSuffix(homeserver, "/")

	var wellKnown struct {
		Homeserver struct {
			BaseURL string `json:"base_url"`
		} `json:"m.homeserver"`
	}
	wk, wkErr := client.R().SetContext(ctx).SetResult(&wellKnown).Get(entered + "/.well-known/matrix/client")
	if wkErr == nil && !wk.IsError() && wellKnown.Homeserver.BaseURL != "" {
		base = strings.TrimSuffix(wellKnown.Homeserver.BaseURL, "/")
		detail = fmt.Sprintf("entered: %s → client API: %s", entered, base)
		if err := probeHomeserverClientAPI(ctx, client, base); err != nil {
			return "", detail, err
		}
		return base, detail, nil
	}

	detail = fmt.Sprintf("using %s (no .well-known)", entered)
	if wkErr != nil {
		detail += "; .well-known: " + wkErr.Error()
	} else if wk.IsError() {
		detail += fmt.Sprintf("; .well-known: HTTP %d", wk.StatusCode())
	}

	if err := probeHomeserverClientAPI(ctx, client, entered); err != nil {
		return "", detail, err
	}
	return entered, detail, nil
}

// probeHomeserverClientAPI checks that the Matrix Client-Server API responds.
func probeHomeserverClientAPI(ctx context.Context, client *resty.Client, base string) error {
	base = strings.TrimSuffix(strings.TrimSpace(base), "/")
	if base == "" {
		return fmt.Errorf("homeserver url is empty")
	}
	r, err := client.R().SetContext(ctx).Get(base + "/_matrix/client/versions")
	if err != nil {
		return fmt.Errorf("cannot reach homeserver at %s: %w", base, err)
	}
	if r.IsError() {
		return fmt.Errorf("homeserver at %s returned HTTP %d", base, r.StatusCode())
	}
	return nil
}

func checkJoinedRoom(ctx context.Context, client *resty.Client, base, token, roomID string) (ok bool, failMsg, detail string, err error) {
	var joined struct {
		JoinedRooms []string `json:"joined_rooms"`
		ErrCode     string   `json:"errcode"`
		Error       string   `json:"error"`
	}
	r, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&joined).
		Get(base + "/_matrix/client/v3/joined_rooms")
	if err != nil {
		return false, "", "", fmt.Errorf("joined_rooms: %w", err)
	}
	if r.IsError() {
		msg := joined.Error
		if msg == "" {
			msg = r.String()
		}
		return false, "", "", fmt.Errorf("joined_rooms (%d): %s", r.StatusCode(), msg)
	}
	for _, id := range joined.JoinedRooms {
		if id == roomID {
			return true, "Bot/user is in room " + roomID, fmt.Sprintf("found among %d joined rooms", len(joined.JoinedRooms)), nil
		}
	}
	return false,
		fmt.Sprintf("not in room %s — invite the bot and accept the invite", roomID),
		fmt.Sprintf("joined rooms: %d", len(joined.JoinedRooms)),
		nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// ConfigFromMap builds NotifierConfig from access credential JSON.
func ConfigFromMap(m map[string]any) (*NotifierConfig, error) {
	if m == nil {
		return nil, fmt.Errorf("config is empty")
	}
	cfg := &NotifierConfig{}
	if v, ok := m["homeserverUrl"].(string); ok {
		cfg.HomeserverUrl = v
	}
	if v, ok := m["authMode"].(string); ok {
		cfg.AuthMode = v
	}
	if v, ok := m["accessToken"].(string); ok {
		cfg.AccessToken = v
	}
	if v, ok := m["sessionAccessToken"].(string); ok {
		cfg.SessionAccessToken = v
	}
	if v, ok := m["sessionDeviceId"].(string); ok {
		cfg.SessionDeviceId = v
	}
	if v, ok := m["userId"].(string); ok {
		cfg.UserId = v
	}
	if v, ok := m["password"].(string); ok {
		cfg.Password = v
	}
	if v, ok := m["roomId"].(string); ok {
		cfg.RoomId = v
	}
	if strings.TrimSpace(cfg.HomeserverUrl) == "" {
		return nil, fmt.Errorf("homeserver url is required")
	}
	return cfg, nil
}
