package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/internal/tools/smtp"
	"github.com/certimate-go/certimate/pkg/core/notifier"
	"github.com/microcosm-cc/bluemonday"
)

type NotifierConfig struct {
	// SMTP 服务器地址。
	SmtpHost string `json:"smtpHost"`
	// SMTP 服务器端口。
	// 零值时根据是否启用 TLS 决定。
	SmtpPort int32 `json:"smtpPort"`
	// 是否启用 TLS。
	SmtpTls bool `json:"smtpTls"`
	// 用户名。
	Username string `json:"username"`
	// 密码。
	Password string `json:"password"`
	// 发件人邮箱。
	SenderAddress string `json:"senderAddress"`
	// 发件人显示名称。
	SenderName string `json:"senderName,omitempty"`
	// 收件人邮箱。
	ReceiverAddress string `json:"receiverAddress"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
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
	// HTML安全过滤
	safeHtml, err := n.sanitizeHtml(message)
	if err != nil {
		return nil, fmt.Errorf("invalid html content: %w", err)
	}

	clientCfg := smtp.NewDefaultConfig()
	clientCfg.Host = n.config.SmtpHost
	clientCfg.Port = int(n.config.SmtpPort)
	clientCfg.Username = n.config.Username
	clientCfg.Password = n.config.Password
	clientCfg.UseSsl = n.config.SmtpTls
	clientCfg.SkipTlsVerify = n.config.AllowInsecureConnections
	client, err := smtp.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	defer client.Close()

	msg := smtp.NewMessage()
	msg.Subject(subject)
	// Html正文
	msg.SetBodyString(smtp.MIMETypeTextHTML, safeHtml)

	// 增加纯文本fallback
	plainFallback := bluemonday.StrictPolicy().Sanitize(safeHtml)
	msg.AddAlternativeString(smtp.MIMETypeTextPlain, plainFallback)
	if n.config.SenderName == "" {
		msg.From(n.config.SenderAddress)
	} else {
		msg.FromFormat(n.config.SenderName, n.config.SenderAddress)
	}
	msg.To(n.config.ReceiverAddress)

	if err := client.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to send mail: %w", err)
	}

	return &notifier.NotifyResult{}, nil
}

func (n *Notifier) sanitizeHtml(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("html content is empty")
	}

	safe := bluemonday.UGCPolicy().Sanitize(input)

	if strings.TrimSpace(safe) == "" {
		return "", fmt.Errorf("html content removed by sanitizer")
	}
	return safe, nil
}
