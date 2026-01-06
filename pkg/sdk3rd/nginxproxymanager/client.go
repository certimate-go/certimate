package nginxproxymanager

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	identity string
	secret   string

	jwtToken    string
	jwtTokenMtx sync.Mutex

	client *resty.Client
}

func NewClient(serverUrl, identity, secret string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if identity == "" {
		return nil, fmt.Errorf("sdkerr: unset identity")
	}
	if secret == "" {
		return nil, fmt.Errorf("sdkerr: unset secret")
	}

	client := &Client{
		identity: identity,
		secret:   secret,
	}
	client.client = resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")+"/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.jwtToken != "" {
				req.Header.Set("Authorization", "Bearer "+client.jwtToken)
			}

			return nil
		})

	return client, nil
}

func NewClientWithJwtToken(serverUrl, jwtToken string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if jwtToken == "" {
		return nil, fmt.Errorf("sdkerr: unset jwtToken")
	}

	client := &Client{
		jwtToken: jwtToken,
	}
	client.client = resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")+"/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.jwtToken != "" {
				req.Header.Set("Authorization", "Bearer "+client.jwtToken)
			}

			return nil
		})
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

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.client.R()
	req.Method = method
	req.URL = path
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	// WARN:
	//   PLEASE DO NOT USE `req.SetResult` or `req.SetError` HERE! USE `doRequestWithResult` INSTEAD.

	resp, err := req.Send()
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) doRequestWithResult(req *resty.Request, res interface{}) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.doRequest(req)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &res)
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		var errRes *apiResponseBase
		if err := json.Unmarshal(resp.Body(), &errRes); err == nil {
			if terror := errRes.GetError(); terror != "" {
				return resp, fmt.Errorf("sdkerr: error='%s'", terror)
			}
		}

		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}

func (c *Client) ensureJwtTokenExists() error {
	c.jwtTokenMtx.Lock()
	defer c.jwtTokenMtx.Unlock()
	if c.jwtToken != "" {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, "/tokens")
	if err != nil {
		return err
	} else {
		httpreq.SetBody(map[string]string{
			"identity": c.identity,
			"secret":   c.secret,
		})
	}

	type tokensResponse struct {
		apiResponseBase
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}

	result := &tokensResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else if terror := result.GetError(); terror != "" {
		return fmt.Errorf("sdkerr: failed to create npm token: error='%s'", terror)
	} else {
		c.jwtToken = result.Token
	}

	return nil
}
