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
	client *resty.Client
}

func NewClient(serverUrl string, userId string, accessToken string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if userId == "" {
		return nil, fmt.Errorf("sdkerr: unset userId")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("sdkerr: unset accessToken")
	}

	baseUrl, _ := resolveBaseUrl(serverUrl)
	if baseUrl == "" {
		baseUrl = serverUrl
	}

	client := &Client{}
	client.client = resty.New().
		SetBaseURL(strings.TrimSuffix(baseUrl, "/")).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetAuthToken(accessToken)
	if err := client.probeVersions(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.client.SetTLSClientConfig(config)
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
		return strings.TrimSuffix(wkJSON.Homeserver.BaseURL, "/"), nil
	} else {
		return serverUrl, nil
	}
}

func newTransactionId() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("certimate_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
