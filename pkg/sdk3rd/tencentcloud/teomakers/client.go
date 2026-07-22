// A simple SDK client for Tencent Cloud EdgeOne Pages.
// API documentation: https://docs.edgeone.site/
package teomakers

import (
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
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if options.ApiToken == "" {
		return nil, fmt.Errorf("sdkerr: unset apiToken")
	}

	httper := resty.New().
		SetBaseURL("https://pages-api.cloud.tencent.com/v1").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+options.ApiToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent)

	return &Client{rc: httper}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(params any, xAction string) (*resty.Request, error) {
	if xAction == "" {
		return nil, fmt.Errorf("sdkerr: unset action")
	}

	paramsMap := map[string]any{}
	paramsMap["Action"] = xAction
	if params != nil {
		jsonb, _ := json.Marshal(params)
		json.Unmarshal(jsonb, &paramsMap)
		if paramsMap["Action"] != xAction {
			return nil, fmt.Errorf("sdkerr: bad request: action mismatch: expected '%s', got '%s'", xAction, paramsMap["Action"])
		}
	}

	req := c.rc.R()
	req.Method = http.MethodPost
	req.URL = "/"
	req.SetBody(paramsMap)

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
				return resp, fmt.Errorf("sdkerr: api error: code='%d', message='%s'", rCode, res.GetMessage())
			}
		}
	}

	return resp, nil
}
