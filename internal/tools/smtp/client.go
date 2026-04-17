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
		return nil, fmt.Errorf("smtp: %w", err)
	}

	return &Client{cli: client}, nil
}

func (c *Client) RawClient() *mail.Client {
	return c.cli
}

func (c *Client) Close() error {
	err := c.cli.Close()
	if err != nil {
		return fmt.Errorf("smtp: %w", err)
	}

	return nil
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

	if config.Port == 0 {
		if config.UseSsl {
			clientOptions = append(clientOptions, mail.WithPort(mail.DefaultPortSSL))
		} else {
			clientOptions = append(clientOptions, mail.WithPort(mail.DefaultPort))
		}
	} else {
		clientOptions = append(clientOptions, mail.WithPort(config.Port))
	}

	if config.UseSsl {
		tlsConfig := xtls.NewCompatibleConfig()
		if config.SkipTlsVerify {
			tlsConfig.InsecureSkipVerify = true
		} else {
			tlsConfig.ServerName = config.Host
		}

		clientOptions = append(clientOptions, mail.WithSSL())
		clientOptions = append(clientOptions, mail.WithTLSConfig(tlsConfig))
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.TLSMandatory))
	} else {
		clientOptions = append(clientOptions, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(config.Host, clientOptions...)
	if err != nil {
		return nil, err
	}

	client.ErrorHandlerRegistry.RegisterHandler("smtp.qq.com", "QUIT", &wQQMailQuitErrorHandler{})

	return client, nil
}
