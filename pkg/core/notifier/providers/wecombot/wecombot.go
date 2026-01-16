package wecombot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// 企业微信机器人 Webhook 地址。
	WebhookUrl string `json:"webhookUrl"`
	// 自定义消息数据。
	// 选填。
	CustomPayload string `json:"customPayload,omitempty"`
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
	webhookUrl, err := url.Parse(n.config.WebhookUrl)
	if err != nil {
		return nil, fmt.Errorf("dingtalk api error: invalid webhook url: %w", err)
	} else {
		const hostname = "qyapi.weixin.qq.com"
		if webhookUrl.Hostname() != hostname {
			n.logger.Warn(fmt.Sprintf("the webhook url hostname is not '%s', please make sure it is correct", hostname))
		}
	}

	var webhookData map[string]any
	if n.config.CustomPayload == "" {
		webhookData = map[string]any{
			"msgtype": "text",
			"text": map[string]string{
				"content": subject + "\n\n" + message,
			},
		}
	} else {
		err = json.Unmarshal([]byte(n.config.CustomPayload), &webhookData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
		}

		replaceJsonValueRecursively(webhookData, "${CERTIMATE_NOTIFIER_SUBJECT}", subject)
		replaceJsonValueRecursively(webhookData, "${CERTIMATE_NOTIFIER_MESSAGE}", message)

		replaceJsonValueRecursively(webhookData, "${SUBJECT}", subject)
		replaceJsonValueRecursively(webhookData, "${MESSAGE}", message)
	}

	// REF: https://developer.work.weixin.qq.com/document/path/91770
	var result struct {
		ErrorCode    int    `json:"errcode"`
		ErrorMessage string `json:"errmsg"`
	}
	req := n.httpClient.R().
		SetContext(ctx).
		SetBody(webhookData)
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

func replaceJsonValueRecursively(data interface{}, oldStr, newStr string) interface{} {
	switch v := data.(type) {
	case map[string]any:
		for k, val := range v {
			v[k] = replaceJsonValueRecursively(val, oldStr, newStr)
		}
	case []any:
		for i, val := range v {
			v[i] = replaceJsonValueRecursively(val, oldStr, newStr)
		}
	case []string:
		for i, s := range v {
			var val interface{} = s
			var newVal interface{} = replaceJsonValueRecursively(val, oldStr, newStr)
			v[i] = newVal.(string)
		}
	case string:
		return strings.ReplaceAll(v, oldStr, newStr)
	}
	return data
}
