package smtp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"

	xtls "github.com/certimate-go/certimate/pkg/utils/tls"
)

type Client struct {
	cli *mail.Client
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of SMTP client is nil")
	}

	client, err := createSmtpClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{cli: client}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Send(ctx context.Context, msg *Message) error {
	if err := c.cli.DialAndSendWithContext(ctx, msg); err != nil {
		errShouldBeIgnored := false

		// REF: https://github.com/wneessen/go-mail/issues/463
		var sendErr *mail.SendError
		if errors.As(err, &sendErr) {
			if sendErr.Reason == mail.ErrSMTPReset {
				errShouldBeIgnored = true
			}
		}

		if !errShouldBeIgnored {
			return fmt.Errorf("smtp: %w", err)
		}
	}

	return nil
}

func createSmtpClient(config *Config) (*mail.Client, error) {
	clientOptions := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(config.Username),
		mail.WithPassword(config.Password),
		mail.WithTimeout(time.Second * 30),
	}

	// Determine the actual port to use
	port := config.Port
	if port == 0 {
		if config.UseSsl {
			port = mail.DefaultPortSSL // 465
		} else {
			port = mail.DefaultPort // 25
		}
	}
	clientOptions = append(clientOptions, mail.WithPort(port))

	if config.UseSsl {
		tlsConfig := xtls.NewCompatibleConfig()
		if config.SkipTlsVerify {
			tlsConfig.InsecureSkipVerify = true
		} else {
			tlsConfig.ServerName = config.Host
		}

		clientOptions = append(clientOptions, mail.WithTLSConfig(tlsConfig))

		// Port 465 uses implicit TLS (SMTPS), while port 587/25 uses explicit TLS (STARTTLS).
		// For implicit TLS, we need WithSSL(); for STARTTLS, we only need TLSMandatory policy.
		if port == 465 {
			clientOptions = append(clientOptions, mail.WithSSL())
		}
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.TLSMandatory))
	} else {
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(config.Host, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("smtp: %w", err)
	}

	client.ErrorHandlerRegistry.RegisterHandler("smtp.qq.com", "QUIT", &wQQMailQuitErrorHandler{})

	return client, nil
}
