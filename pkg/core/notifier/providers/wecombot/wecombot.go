package wecombot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// 企业微信机器人 Webhook 地址。
	WebhookUrl string `json:"webhookUrl"`
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
	webhookUrl, err := url.Parse(n.config.WebhookUrl)
	if err != nil {
		return nil, fmt.Errorf("dingtalk api error: invalid webhook url: %w", err)
	} else {
		const hostname = "qyapi.weixin.qq.com"
		if webhookUrl.Hostname() != hostname {
			n.logger.Warn(fmt.Sprintf("the webhook url hostname is not '%s', please make sure it is correct", hostname))
		}
	}

	// REF: https://developer.work.weixin.qq.com/document/path/91770
	var result struct {
		ErrorCode    int    `json:"errcode"`
		ErrorMessage string `json:"errmsg"`
	}
	req := n.httpClient.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"msgtype": "text",
			"text": map[string]string{
				"content": subject + "\n\n" + message,
			},
		})
	resp, err := req.Post(webhookUrl.String())
	if err != nil {
		return nil, fmt.Errorf("wecom api error: failed to send request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("wecom api error: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	} else if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("wecom api error: %w (resp: %s)", err, string(resp.Body()))
	} else if result.ErrorCode != 0 {
		return nil, fmt.Errorf("wecom api error: errcode='%d', errmsg='%s'", result.ErrorCode, result.ErrorMessage)
	}

	return &notifier.NotifyResult{}, nil
}
