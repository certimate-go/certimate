package matrix

import (
	"context"
	"fmt"
	"strings"
)

const (
	defaultTestSubject = "Certimate test notification"
	defaultTestMessage = "This is a test message from credential settings. If you see this in the room, sending works."
)

// SendTestMessage posts a test notification to cfg.RoomId using the configured auth.
// On password login it may set a pending session on cfg (same as Notify).
func SendTestMessage(ctx context.Context, cfg *NotifierConfig, subject, message string) error {
	if cfg == nil {
		return fmt.Errorf("matrix: config is nil")
	}
	if strings.TrimSpace(cfg.RoomId) == "" {
		return fmt.Errorf("matrix: room id is required")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = defaultTestSubject
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = defaultTestMessage
	}

	n, err := NewNotifier(cfg)
	if err != nil {
		return err
	}
	_, err = n.Notify(ctx, subject, message)
	return err
}
