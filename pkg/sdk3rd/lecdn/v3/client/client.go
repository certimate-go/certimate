package client

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
)

type Client struct {
	username string
	password string

	accessToken    string
	accessTokenMtx sync.Mutex

	client *resty.Client
}

func NewClient(serverUrl, username, password string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if username == "" {
		return nil, fmt.Errorf("sdkerr: unset username")
	}
	if password == "" {
		return nil, fmt.Errorf("sdkerr: unset password")
	}

	client := &Client{
		username: username,
		password: password,
	}
	client.client = resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")+"/prod-api").
		SetHeader("User-Agent", "certimate").
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+client.accessToken)
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

func (c *Client) doRequestWithResult(req *resty.Request, res apiResponse) (*resty.Response, error) {
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
			if tcode := res.GetCode(); tcode != 200 {
				return resp, fmt.Errorf("sdkerr: code='%d', message='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureAccessTokenExists() error {
	c.accessTokenMtx.Lock()
	defer c.accessTokenMtx.Unlock()
	if c.accessToken != "" {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, "/auth/login")
	if err != nil {
		return err
	} else {
		httpreq.SetBody(map[string]string{
			"email":    c.username,
			"username": c.username,
			"password": c.password,
		})
	}

	type loginResponse struct {
		apiResponseBase
		Data *struct {
			UserId   int64  `json:"user_id"`
			Username string `json:"username"`
			Token    string `json:"token"`
		} `json:"data,omitempty"`
	}

	result := &loginResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else {
		c.accessToken = result.Data.Token
	}

	return nil
}
