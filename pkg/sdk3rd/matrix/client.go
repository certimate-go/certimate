package matrix

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

// Client calls the Matrix Client-Server API.
// Клиент для Matrix Client-Server API.
// REF: https://spec.matrix.org/latest/client-server-api/
type Client struct {
	enteredURL string
	baseURL    string
	http       *resty.Client
}

// NewClient builds an HTTP client for the given homeserver URL (scheme added if missing).
// Создаёт HTTP-клиент для указанного URL homeserver (схема https добавляется при отсутствии).
func NewClient(homeserverURL string) (*Client, error) {
	entered := strings.TrimSpace(homeserverURL)
	if entered == "" {
		return nil, fmt.Errorf("sdkerr: unset homeserver url")
	}
	if !strings.HasPrefix(entered, "http://") && !strings.HasPrefix(entered, "https://") {
		entered = "https://" + entered
	}
	entered = strings.TrimSuffix(entered, "/")
	if _, err := url.Parse(entered); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid homeserver url: %w", err)
	}

	http := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetTimeout(60 * time.Second)

	return &Client{
		enteredURL: entered,
		http:       http,
	}, nil
}

// ResolveBaseURL discovers the Client-Server API base URL.
// Определяет базовый URL Client-Server API (well-known или введённый адрес).
// REF: https://spec.matrix.org/latest/client-server-api/#getwell-knownmatrixclient
func (c *Client) ResolveBaseURL() (string, error) {
	if c.baseURL != "" {
		return c.baseURL, nil
	}

	var wellKnown struct {
		Homeserver struct {
			BaseURL string `json:"base_url"`
		} `json:"m.homeserver"`
	}
	wk, err := c.http.R().SetResult(&wellKnown).Get(c.enteredURL + "/.well-known/matrix/client")
	if err == nil && !wk.IsError() && strings.TrimSpace(wellKnown.Homeserver.BaseURL) != "" {
		c.baseURL = strings.TrimSuffix(wellKnown.Homeserver.BaseURL, "/")
	} else {
		c.baseURL = c.enteredURL
	}

	if err := c.probeVersions(); err != nil {
		return "", err
	}
	return c.baseURL, nil
}
