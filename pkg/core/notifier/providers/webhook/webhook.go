package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/core/notifier"
)

type NotifierConfig struct {
	// Webhook URL。
	WebhookUrl string `json:"webhookUrl"`
	// Webhook 回调数据（application/json 或 application/x-www-form-urlencoded 格式）。
	WebhookData string `json:"webhookData,omitempty"`
	// 请求谓词。
	// 零值时默认值 "POST"。
	Method string `json:"method,omitempty"`
	// 请求标头。
	Headers map[string]string `json:"headers,omitempty"`
	// 请求超时（单位：秒）。
	// 零值时默认值 30。
	Timeout int `json:"timeout,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
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
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second)
	if config.Timeout > 0 {
		client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	}
	if config.AllowInsecureConnections {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

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
	// 处理 Webhook URL
	webhookUrl, err := url.Parse(n.config.WebhookUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook url: %w", err)
	} else if webhookUrl.Scheme != "http" && webhookUrl.Scheme != "https" {
		return nil, fmt.Errorf("unsupported webhook url scheme '%s'", webhookUrl.Scheme)
	}

	// 处理 Webhook 请求谓词
	webhookMethod := strings.ToUpper(n.config.Method)
	if webhookMethod == "" {
		webhookMethod = http.MethodPost
	} else if webhookMethod != http.MethodGet &&
		webhookMethod != http.MethodPost &&
		webhookMethod != http.MethodPut &&
		webhookMethod != http.MethodPatch &&
		webhookMethod != http.MethodDelete {
		return nil, fmt.Errorf("unsupported webhook request method '%s'", webhookMethod)
	}

	// 处理 Webhook 请求标头
	webhookHeaders := make(http.Header)
	for k, v := range n.config.Headers {
		webhookHeaders.Set(k, v)
	}

	// 处理 Webhook 请求内容类型
	const CONTENT_TYPE_JSON = "application/json"
	const CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
	const CONTENT_TYPE_MULTIPART = "multipart/form-data"
	webhookContentType := webhookHeaders.Get("Content-Type")
	if webhookContentType == "" {
		webhookContentType = CONTENT_TYPE_JSON
		webhookHeaders.Set("Content-Type", CONTENT_TYPE_JSON)
	} else if strings.HasPrefix(webhookContentType, CONTENT_TYPE_JSON) &&
		strings.HasPrefix(webhookContentType, CONTENT_TYPE_FORM) &&
		strings.HasPrefix(webhookContentType, CONTENT_TYPE_MULTIPART) {
		return nil, fmt.Errorf("unsupported webhook content type '%s'", webhookContentType)
	}

	// 处理 Webhook 请求数据
	var webhookData interface{}
	if n.config.WebhookData == "" {
		webhookData = map[string]string{
			"subject": subject,
			"message": message,
		}
	} else {
		err = json.Unmarshal([]byte(n.config.WebhookData), &webhookData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
		}

		if webhookMethod == http.MethodGet || webhookContentType == CONTENT_TYPE_FORM || webhookContentType == CONTENT_TYPE_MULTIPART {
			temp := make(map[string]string)
			jsonb, err := json.Marshal(webhookData)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
			} else if err := json.Unmarshal(jsonb, &temp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook data: %w", err)
			} else {
				webhookData = temp
			}
		}
	}

	// 替换变量值
	replaceJsonValueRecursively(webhookData, "${CERTIMATE_NOTIFIER_SUBJECT}", subject)
	replaceJsonValueRecursively(webhookData, "${CERTIMATE_NOTIFIER_MESSAGE}", message)

	// 兼容旧版变量
	replaceJsonValueRecursively(webhookData, "${SUBJECT}", subject)
	replaceJsonValueRecursively(webhookData, "${MESSAGE}", message)

	// 生成请求
	// 其中 GET 请求需转换为查询参数
	req := n.httpClient.R().SetContext(ctx).SetHeaderMultiValues(webhookHeaders)
	req.URL = webhookUrl.String()
	req.Method = webhookMethod
	if webhookMethod == http.MethodGet {
		req.SetQueryParams(webhookData.(map[string]string))
	} else {
		switch webhookContentType {
		case CONTENT_TYPE_JSON:
			req.SetBody(webhookData)
		case CONTENT_TYPE_FORM:
			req.SetFormData(webhookData.(map[string]string))
		case CONTENT_TYPE_MULTIPART:
			req.SetMultipartFormData(webhookData.(map[string]string))
		}
	}

	// 发送请求
	resp, err := req.Send()
	if err != nil {
		return nil, fmt.Errorf("webhook error: failed to send request: %w", err)
	} else if resp.IsError() {
		return nil, fmt.Errorf("webhook error: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	n.logger.Debug("webhook responded", slog.String("response", resp.String()))

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
	case string:
		return strings.ReplaceAll(v, oldStr, newStr)
	}
	return data
}
