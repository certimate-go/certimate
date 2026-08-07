// A simple SDK client for NginxProxyManager.
// API documentation: https://github.com/NginxProxyManager/nginx-proxy-manager/discussions/3265
package nginxproxymanager

import (
	"context"
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
	username string
	password string

	token   string
	tokenMu sync.Mutex

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
	if opts.JwtToken == "" && (opts.Username == "" || opts.Password == "") {
		return nil, fmt.Errorf("sdkerr: unset password or jwtToken")
	}

	client := &Client{
		username: opts.Username,
		password: opts.Password,
		token:    opts.JwtToken,
	}
	client.rc = resty.New().
		SetBaseURL(strings.TrimSuffix(serverUrl, "/")+"/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.token != "" {
				req.Header.Set("Authorization", "Bearer "+client.token)
			}

			return nil
		})

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

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.rc.R()
	req.Method = method
	req.URL = path

	// WARN:
	//   DO NOT CALL `req.SetResult` or `req.SetError` AGAIN! USE `doRequestWithResult` INSTEAD.
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := req.Send()
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) doRequestWithResult(req *resty.Request, res any) (*resty.Response, error) {
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
		var errRes *sdkResponseBase
		if err := json.Unmarshal(resp.Body(), &errRes); err == nil {
			if rError := errRes.GetError(); rError != "" {
				return resp, fmt.Errorf("sdkerr: api error: error='%s'", rError)
			}
		}

		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, "/tokens")
	if err != nil {
		return err
	} else {
		httpreq.SetBody(map[string]string{
			"identity": c.username,
			"secret":   c.password,
		})
		httpreq.SetContext(ctx)
	}

	type tokensResponse struct {
		sdkResponseBase
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}

	result := &tokensResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else if rError := result.GetError(); rError != "" {
		return fmt.Errorf("sdkerr: auth error: error='%s'", rError)
	} else {
		if result.Token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		c.token = result.Token
	}

	return nil
}
