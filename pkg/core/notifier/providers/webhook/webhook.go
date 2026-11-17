package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/core"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type (
	Provider     = core.Notifier
	NotifyResult = core.NotifierNotifyResult
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

var _ Provider = (*Notifier)(nil)

const (
	contentTypeJson      = "application/json"
	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeMultipart = "multipart/form-data"
)

var allowedContentTypes = map[string]bool{
	contentTypeJson:      true,
	contentTypeForm:      true,
	contentTypeMultipart: true,
}

var allowedMethods = map[string]bool{
	http.MethodGet:    true,
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

func NewNotifier(config *NotifierConfig) (*Notifier, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the notifier provider is nil")
	}

	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		AddRetryCondition(func(resp *resty.Response, _ error) bool {
			return resp == nil || resp.StatusCode() >= 500
		})
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

func (n *Notifier) Notify(ctx context.Context, subject string, message string) (*NotifyResult, error) {
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
	} else if !allowedMethods[webhookMethod] {
		return nil, fmt.Errorf("unsupported webhook request method '%s'", webhookMethod)
	}

	// 处理 Webhook 请求标头
	webhookHeaders := make(http.Header)
	for k, v := range n.config.Headers {
		webhookHeaders.Set(k, v)
	}

	// 处理 Webhook 请求内容类型
	webhookContentType := webhookHeaders.Get("Content-Type")
	if webhookContentType == "" {
		webhookContentType = contentTypeJson
		webhookHeaders.Set("Content-Type", contentTypeJson)
	} else if mediaType, _, err := mime.ParseMediaType(webhookContentType); err != nil || !allowedContentTypes[mediaType] {
		return nil, fmt.Errorf("unsupported webhook content type '%s'", webhookContentType)
	}

	// 处理 Webhook 请求数据
	var webhookData any
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

		if webhookMethod == http.MethodGet || webhookContentType == contentTypeForm || webhookContentType == contentTypeMultipart {
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
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_NOTIFIER_SUBJECT}", subject)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${CERTIMATE_NOTIFIER_MESSAGE}", message)

	// 兼容旧版变量
	// TODO: remove in future version
	xmaps.DeepReplaceValueUnsafe(webhookData, "${SUBJECT}", subject)
	xmaps.DeepReplaceValueUnsafe(webhookData, "${MESSAGE}", message)

	// 生成请求
	// 其中 GET 请求需转换为查询参数
	req := n.httpClient.R().SetContext(ctx).SetHeaderMultiValues(webhookHeaders)
	req.URL = webhookUrl.String()
	req.Method = webhookMethod
	if webhookMethod == http.MethodGet {
		req.SetQueryParams(webhookData.(map[string]string))
	} else {
		switch webhookContentType {
		case contentTypeJson:
			req.SetBody(webhookData)
		case contentTypeForm:
			req.SetFormData(webhookData.(map[string]string))
		case contentTypeMultipart:
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

	return &NotifyResult{}, nil
}
