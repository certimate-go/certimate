package matrix

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SendText posts an m.room.message event (m.text).
// Отправляет текстовое сообщение в комнату (m.room.message, msgtype m.text).
// REF: https://spec.matrix.org/latest/client-server-api/#put_matrixclientv3roomsroomidsendeventtypetxnid
func (c *Client) SendText(ctx context.Context, accessToken, roomID, body string) error {
	if strings.TrimSpace(accessToken) == "" {
		return fmt.Errorf("sdkerr: unset access token")
	}
	if strings.TrimSpace(roomID) == "" {
		return fmt.Errorf("sdkerr: unset room id")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("sdkerr: unset message body")
	}

	base, err := c.ResolveBaseURL()
	if err != nil {
		return err
	}

	txnID := newTxnID()
	path := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		base, url.PathEscape(roomID), url.PathEscape(txnID))

	payload := map[string]any{
		"msgtype": "m.text",
		"body":    body,
	}

	r, err := c.http.R().
		SetContext(ctx).
		SetAuthToken(accessToken).
		SetBody(payload).
		Put(path)
	if err != nil {
		return fmt.Errorf("sdkerr: send message: %w", err)
	}
	if r.IsError() {
		return fmt.Errorf("sdkerr: send message: status %d: %s", r.StatusCode(), r.String())
	}
	return nil
}

// newTxnID returns a unique transaction id for the send endpoint (idempotency key).
// Генерирует уникальный txn id для PUT send (ключ идемпотентности).
func newTxnID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("certimate_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
