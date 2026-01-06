package goedge

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
	apiRole     string
	accessKeyId string
	accessKey   string

	accessToken    string
	accessTokenExp time.Time
	accessTokenMtx sync.Mutex

	client *resty.Client
}

func NewClient(serverUrl, apiRole, accessKeyId, accessKey string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if apiRole == "" {
		return nil, fmt.Errorf("sdkerr: unset apiRole")
	}
	if apiRole != "user" && apiRole != "admin" {
		return nil, fmt.Errorf("sdkerr: invalid apiRole")
	}
	if accessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKey")
	}

	client := &Client{
		apiRole:     apiRole,
		accessKeyId: accessKeyId,
		accessKey:   accessKey,
	}
	client.client = resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.accessToken != "" {
				req.Header.Set("X-Edge-Access-Token", client.accessToken)
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
	if c.accessToken != "" && c.accessTokenExp.After(time.Now()) {
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
	}

	type getAPIAccessTokenResponse struct {
		apiResponseBase
		Data *struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data,omitempty"`
	}

	result := &getAPIAccessTokenResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else if code := result.GetCode(); code != 200 {
		return fmt.Errorf("sdkerr: failed to get goedge access token: code='%d', message='%s'", code, result.GetMessage())
	} else {
		c.accessToken = result.Data.Token
		c.accessTokenExp = time.Unix(result.Data.ExpiresAt, 0)
	}

	return nil
}
