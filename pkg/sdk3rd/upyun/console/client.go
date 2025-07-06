package console

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	username string
	password string

	loginCookie    string
	loginCookieMtx sync.Mutex

	client *resty.Client
}

func NewClient(username, password string) (*Client, error) {
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
		SetBaseURL("https://console.upyun.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate").
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.loginCookie != "" {
				req.Header.Set("Cookie", client.loginCookie)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
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
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d, resp: %s", resp.StatusCode(), resp.String())
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
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
		} else {
			tresp := &apiResponseBase{}
			if err := json.Unmarshal(resp.Body(), &tresp); err != nil {
				return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
			} else if tdata := tresp.GetData(); tdata == nil {
				return resp, fmt.Errorf("sdkerr: received empty data")
			} else if terrcode := tdata.GetErrorCode(); terrcode != 0 {
				return resp, fmt.Errorf("sdkerr: code='%d', message='%s'", terrcode, tdata.GetMessage())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureCookieExists() error {
	c.loginCookieMtx.Lock()
	defer c.loginCookieMtx.Unlock()
	if c.loginCookie != "" {
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
	}

	type signinResponse struct {
		apiResponseBase
		Data *struct {
			apiResponseBaseData
			Result bool `json:"result"`
		} `json:"data,omitempty"`
	}

	result := &signinResponse{}
	httpresp, err := c.doRequestWithResult(httpreq, result)
	if err != nil {
		return err
	} else if !result.Data.Result {
		return errors.New("sdkerr: failed to signin upyun console")
	} else {
		c.loginCookie = httpresp.Header().Get("Set-Cookie")
	}

	return nil
}
