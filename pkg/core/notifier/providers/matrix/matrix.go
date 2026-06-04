package matrix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	matrixsdk "github.com/certimate-go/certimate/pkg/sdk3rd/matrix"
	"github.com/certimate-go/certimate/pkg/core/notifier"
)

// NotifierConfig holds Matrix credentials and the default notification room.
// Параметры Matrix: homeserver, пользователь, токен и комната для уведомлений.
type NotifierConfig struct {
	// Homeserver URL (Element web URL or Matrix homeserver base URL).
	// URL homeserver (адрес Element или базовый URL Matrix).
	HomeserverUrl string `json:"homeserverUrl"`
	// User ID (MXID), e.g. @bot:example.org.
	// Идентификатор пользователя (MXID), например @bot:example.org.
	UserId string `json:"userId"`
	// Access token from the homeserver (bot or user).
	// Access token с homeserver (бот или пользователь).
	AccessToken string `json:"accessToken"`
	// Default room ID (!room:server) for notifications.
	// ID комнаты по умолчанию (!room:server) для уведомлений.
	RoomId string `json:"roomId,omitempty"`
}

type Notifier struct {
	config *NotifierConfig
	logger *slog.Logger
}

var _ notifier.Provider = (*Notifier)(nil)

// NewNotifier creates a Matrix notification provider from config.
// Создаёт провайдер уведомлений Matrix по конфигурации.
func NewNotifier(config *NotifierConfig) (*Notifier, error) {
	if config == nil {
		return nil, errors.New("the configuration of the notifier provider is nil")
	}
	return &Notifier{
		config: config,
		logger: slog.Default(),
	}, nil
}

func (n *Notifier) SetLogger(logger *slog.Logger) {
	if logger == nil {
		n.logger = slog.New(slog.DiscardHandler)
	} else {
		n.logger = logger
	}
}

// Notify sends subject and message to the configured room via pkg/sdk3rd/matrix.
// Отправляет subject и message в комнату через pkg/sdk3rd/matrix.
func (n *Notifier) Notify(ctx context.Context, subject string, message string) (*notifier.NotifyResult, error) {
	roomID := strings.TrimSpace(n.config.RoomId)
	if roomID == "" {
		return nil, errors.New("matrix: room id is required")
	}

	token := strings.TrimSpace(n.config.AccessToken)
	if token == "" {
		return nil, errors.New("matrix: access token is required")
	}

	client, err := matrixsdk.NewClient(n.config.HomeserverUrl)
	if err != nil {
		return nil, fmt.Errorf("matrix: %w", err)
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
		slog.String("homeserver", n.config.HomeserverUrl),
		slog.String("userId", strings.TrimSpace(n.config.UserId)),
		slog.String("roomId", roomID),
	)

	if err := client.SendText(ctx, token, roomID, body); err != nil {
		return nil, fmt.Errorf("matrix: %w", err)
	}

	n.logger.Info("matrix: message sent")
	return &notifier.NotifyResult{}, nil
}
