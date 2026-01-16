package dingtalkbot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// 钉钉机器人的 Webhook 地址。
	WebhookUrl string `json:"webhookUrl"`
	// 钉钉机器人的 Secret。
	Secret string `json:"secret"`
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
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if config.Secret != "" {
				timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())

				h := hmac.New(sha256.New, []byte(config.Secret))
				h.Write([]byte(fmt.Sprintf("%s\n%s", timestamp, config.Secret)))
				sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

				qs := req.URL.Query()
				qs.Set("timestamp", timestamp)
				qs.Set("sign", sign)
				req.URL.RawQuery = qs.Encode()
			}

			return nil
		})

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
		const hostname = "oapi.dingtalk.com"
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

	// REF: https://open.dingtalk.com/document/development/custom-robots-send-group-messages
	var result struct {
		ErrorCode    int    `json:"errcode"`
		ErrorMessage string `json:"errmsg"`
	}
	req := n.httpClient.R().
		SetContext(ctx).
		SetBody(webhookData)
	resp, err := req.Post(webhookUrl.String())
	if err != nil {
		return nil, fmt.Errorf("dingtalk api error: failed to send request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("dingtalk api error: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	} else if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("dingtalk api error: %w (resp: %s)", err, resp.String())
	} else if result.ErrorCode != 0 {
		return nil, fmt.Errorf("dingtalk api error: errcode='%d', errmsg='%s'", result.ErrorCode, result.ErrorMessage)
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
