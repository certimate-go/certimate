// A mock HTTP client for Upyun Console.
package console

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	username string
	password string

	cookies   string
	cookiesMu sync.Mutex

	rc *resty.Client
}

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if options.Username == "" {
		return nil, fmt.Errorf("sdkerr: unset username")
	}
	if options.Password == "" {
		return nil, fmt.Errorf("sdkerr: unset password")
	}

	client := &Client{
		username: options.Username,
		password: options.Password,
	}
	client.rc = resty.New().
		SetBaseURL("https://console.upyun.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.cookies != "" {
				req.Header.Set("Cookie", client.cookies)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
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
			tresp := &sdkResponseBase{}
			if err := json.Unmarshal(resp.Body(), &tresp); err != nil {
				return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
			} else if tdata := tresp.GetData(); tdata == nil {
				return resp, fmt.Errorf("sdkerr: received empty data")
			} else if terrcode := tdata.GetErrorCode(); terrcode != 0 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', message='%s'", terrcode, tdata.GetMessage())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureCookies(ctx context.Context) error {
	c.cookiesMu.Lock()
	defer c.cookiesMu.Unlock()
	if c.cookies != "" {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, "/accounts/signin/")
	if err != nil {
		return err
	} else {
		httpreq.SetBody(map[string]string{
			"username": c.username,
			"password": c.password,
		})
		httpreq.SetContext(ctx)
	}

	type signinResponse struct {
		sdkResponseBase
		Data *struct {
			sdkResponseBaseData
			Result bool `json:"result"`
		} `json:"data,omitempty"`
	}

	result := &signinResponse{}
	httpresp, err := c.doRequestWithResult(httpreq, result)
	if err != nil {
		return err
	} else if !result.Data.Result {
		return fmt.Errorf("sdkerr: auth error")
	} else {
		cookies := httpresp.Header().Get("Set-Cookie")
		if cookies == "" {
			return fmt.Errorf("sdkerr: auth error: received empty cookies")
		}

		c.cookies = cookies
	}

	return nil
}
