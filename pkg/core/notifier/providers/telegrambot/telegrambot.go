package telegrambot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// Telegram Bot API Token。
	BotToken string `json:"botToken"`
	// Telegram Chat ID。
	ChatId string `json:"chatId"`
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

	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate")

	return &Notifier{
		config:     config,
		logger:     slog.Default(),
		httpClient: client,
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
	// REF: https://core.telegram.org/bots/api#sendmessage
	req := n.httpClient.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"chat_id": n.config.ChatId,
			"text":    subject + "\n" + message,
		})
	resp, err := req.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.config.BotToken))
	if err != nil {
		return nil, fmt.Errorf("telegram api error: failed to send request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("telegram api error: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return &notifier.NotifyResult{}, nil
}
