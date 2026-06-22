package matrix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/certimate-go/certimate/pkg/core/notifier"
	matrixsdk "github.com/certimate-go/certimate/pkg/sdk3rd/matrix"
)

type NotifierConfig struct {
	// Homeserver URL (Element web URL or Matrix homeserver base URL).
	// URL homeserver (адрес Element или базовый URL Matrix).
	ServerUrl string `json:"serverUrl"`
	// User ID (MXID), e.g. @bot:example.org.
	// Идентификатор пользователя (MXID), например @bot:example.org.
	UserId string `json:"userId"`
	// Access token from the homeserver (bot or user).
	// Access token с homeserver (бот или пользователь).
	AccessToken string `json:"accessToken"`
	// Room ID (!room:server) for notifications.
	// ID комнаты (!room:server) для уведомлений.
	RoomId string `json:"roomId"`
}

type Notifier struct {
	config *NotifierConfig
	logger *slog.Logger
}

var _ notifier.Provider = (*Notifier)(nil)

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

func (n *Notifier) Notify(ctx context.Context, subject string, message string) (*notifier.NotifyResult, error) {
	if n.config.RoomId == "" {
		return nil, errors.New("matrix: config `roomId` is required")
	}

	client, err := matrixsdk.NewClient(n.config.ServerUrl,
		matrixsdk.WithUserId(n.config.UserId),
		matrixsdk.WithAccessToken(n.config.AccessToken),
	)
	if err != nil {
		return nil, fmt.Errorf("matrix: %w", err)
	}

	body := fmt.Sprintf("%s\n\n%s", subject, message)
	if err := client.SendTextMessageToRoom(ctx, n.config.RoomId, body); err != nil {
		return nil, fmt.Errorf("matrix: %w", err)
	}

	return &notifier.NotifyResult{}, nil
}
