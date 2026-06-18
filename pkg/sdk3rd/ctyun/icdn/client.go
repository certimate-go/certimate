// A simple SDK client for StateCloud ICDN.
// API documentation: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=112&vid=166
package icdn

import (
	"fmt"
	"time"

	"github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/openapi"
	"github.com/go-resty/resty/v2"
)

const endpoint = "https://icdn-global.ctapi.ctyun.cn"

type Client struct {
	client *openapi.Client
}

func NewClient(optFns ...openapi.OptionsFunc) (*Client, error) {
	client, err := openapi.NewClient(endpoint, optFns...)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	return c.client.NewRequest(method, path)
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	return c.client.DoRequest(req)
}

func (c *Client) doRequestWithResult(req *resty.Request, res sdkResponse) (*resty.Response, error) {
	resp, err := c.client.DoRequestWithResult(req, res)
	if err == nil {
		rStatusCode := res.GetStatusCode()
		rError := res.GetError()
		if rStatusCode != "" && rStatusCode != "100000" {
			return resp, fmt.Errorf("sdkerr: api error: code='%s', message='%s', error='%s', errorMessage='%s'", rStatusCode, res.GetMessage(), rError, res.GetErrorMessage())
		}
	}

	return resp, err
}
