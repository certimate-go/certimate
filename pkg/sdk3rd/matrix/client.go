package matrix

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	rc *resty.Client
}

func NewClient(serverUrl string, optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if opts.AccessToken == "" {
		return nil, fmt.Errorf("sdkerr: unset accessToken")
	}

	baseUrl, _ := resolveBaseUrl(strings.TrimSuffix(serverUrl, "/"))
	if baseUrl == "" {
		baseUrl = serverUrl
	}

	client := &Client{}
	client.rc = resty.New().
		SetBaseURL(strings.TrimSuffix(baseUrl, "/")).
		SetHeader("Authorization", "Bearer "+opts.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
	if err := client.probeVersions(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.rc.SetTLSClientConfig(config)
	return c
}

func resolveBaseUrl(serverUrl string) (string, error) {
	var wkJSON struct {
		Homeserver struct {
			BaseURL string `json:"base_url"`
		} `json:"m.homeserver"`
	}

	_, err := resty.New().R().
		SetResult(&wkJSON).
		Get(serverUrl + "/.well-known/matrix/client")
	if err != nil {
		return "", fmt.Errorf("failed to discovery Matrix Client API: %w", err)
	} else if strings.TrimSpace(wkJSON.Homeserver.BaseURL) != "" {
		return wkJSON.Homeserver.BaseURL, nil
	} else {
		return serverUrl, nil
	}
}

func newTransactionId() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("certimate_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
