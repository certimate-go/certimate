// A simple SDK client for SamWAF.
// API documentation: https://doc.samwaf.com/api/
package samwaf

import (
	"crypto/tls"
	"encoding/json"
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
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if options.ApiKey == "" {
		return nil, fmt.Errorf("sdkerr: unset apiKey")
	}

	httper := resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")+"/api/v1").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetHeader("X-API-Key", options.ApiKey)

	return &Client{rc: httper}, nil
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
		var errRes *sdkResponseBase
		if err := json.Unmarshal(resp.Body(), &errRes); err == nil {
			if tcode := errRes.GetCode(); tcode != 0 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', msg='%s'", tcode, errRes.GetMsg())
			}
		}

		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}
