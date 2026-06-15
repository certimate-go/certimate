package matrix

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// SendText posts an m.room.message event (m.text).
// Отправляет текстовое сообщение в комнату (m.room.message, msgtype m.text).
// REF: https://spec.matrix.org/latest/client-server-api/#put_matrixclientv3roomsroomidsendeventtypetxnid
func (c *Client) SendTextMessageToRoom(ctx context.Context, roomId, msgBody string) error {
	if strings.TrimSpace(roomId) == "" {
		return fmt.Errorf("sdkerr: unset roomId")
	}

	txnId := newTransactionId()
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		url.PathEscape(roomId), url.PathEscape(txnId))

	payload := map[string]any{
		"msgtype": "m.text",
		"body":    msgBody,
	}

	_, err := c.rc.R().
		SetContext(ctx).
		SetBody(payload).
		Put(path)
	if err != nil {
		return fmt.Errorf("sdkerr: api error: %w", err)
	}

	return nil
}
