// A simple SDK client for Asia-ISP CDN.
package asiaispcdn

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	rc *resty.Client
}

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if opts.AccessKeySecret == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeySecret")
	}

	signer := &signer{
		accessKeyId:     opts.AccessKeyId,
		accessKeySecret: opts.AccessKeySecret,
	}
	httper := resty.New().
		SetBaseURL("https://api.asia-isp.com/openapi/v3/stat").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if err := signer.Sign(req); err != nil {
				return fmt.Errorf("sdkerr: sign error: %w", err)
			}

			return nil
		})

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

func (c *Client) newRequest(method string, params any, xAction string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if xAction == "" {
		return nil, fmt.Errorf("sdkerr: unset action")
	}

	req := c.rc.R()
	req.Method = method
	req.SetQueryParam("action", xAction)
	req.SetBody(params)

	// WARN:
	//   DO NOT CALL `req.SetBody` or `req.SetFormData` AGAIN! USE `newRequest` INSTEAD.
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
			if rCode := res.GetCode(); rCode != 0 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', msg='%s'", rCode, res.GetMsg())
			}
		}
	}

	return resp, nil
}
