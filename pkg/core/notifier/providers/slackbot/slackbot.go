package discordbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// Slack Bot API Token。
	BotToken string `json:"botToken"`
	// Slack Channel ID。
	ChannelId string `json:"channelId"`
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
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.BotToken)).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent)

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
	// REF: https://docs.slack.dev/messaging/sending-and-scheduling-messages#publishing
	req := n.httpClient.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"token":   n.config.BotToken,
			"channel": n.config.ChannelId,
			"text":    subject + "\n" + message,
		})
	resp, err := req.Post("https://slack.com/api/chat.postMessage")
	if err != nil {
		return nil, fmt.Errorf("slack api error: failed to send request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("slack api error: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return &notifier.NotifyResult{}, nil
}
