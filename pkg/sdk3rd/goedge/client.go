// A simple SDK client for GoEdge.
// API documentation: https://goedge.cloud/docs/API/List.md
package goedge

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
	apiRole     string
	accessKeyId string
	accessKey   string

	token   string
	tokenAt time.Time
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
	if opts.Role == "" {
		return nil, fmt.Errorf("sdkerr: unset apiRole")
	}
	if opts.Role != "user" && opts.Role != "admin" {
		return nil, fmt.Errorf("sdkerr: invalid apiRole")
	}
	if opts.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if opts.AccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKey")
	}

	client := &Client{
		apiRole:     opts.Role,
		accessKeyId: opts.AccessKeyId,
		accessKey:   opts.AccessKey,
	}
	client.rc = resty.New().
		SetBaseURL(strings.TrimSuffix(serverUrl, "/")).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.token != "" {
				req.Header.Set("X-Edge-Access-Token", client.token)
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

func (c *Client) doRequestWithResult(req *resty.Request, res sdkResponse) (*resty.Response, error) {
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
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		} else {
			if rCode := res.GetCode(); rCode != 200 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', message='%s'", rCode, res.GetMessage())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && c.tokenAt.After(time.Now()) {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, "/APIAccessTokenService/getAPIAccessToken")
	if err != nil {
		return err
	} else {
		httpreq.SetBody(map[string]string{
			"type":        c.apiRole,
			"accessKeyId": c.accessKeyId,
			"accessKey":   c.accessKey,
		})
		httpreq.SetContext(ctx)
	}

	type getAPIAccessTokenResponse struct {
		sdkResponseBase
		Data *struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data,omitempty"`
	}

	result := &getAPIAccessTokenResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else if rCode := result.GetCode(); rCode != 200 {
		return fmt.Errorf("sdkerr: auth error: code='%d', message='%s'", rCode, result.GetMessage())
	} else {
		if result.Data == nil || result.Data.Token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		tokenAt := time.Unix(result.Data.ExpiresAt, 0)
		if tokenAt.IsZero() {
			return fmt.Errorf("sdkerr: auth error: received invalid token expiration")
		}

		c.token = result.Data.Token
		c.tokenAt = tokenAt
	}

	return nil
}
