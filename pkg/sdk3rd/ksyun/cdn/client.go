// A simple SDK client for KingsoftCloud CDN.
// API documentation: https://apiexplorer.ksyun.com/#/api/home
package cdn

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	common "github.com/certimate-go/certimate/pkg/sdk3rd/ksyun/zz-shared-common"
)

const (
	service  = "cdn"
	endpoint = "https://" + service + ".api.ksyun.com"
)

type Client struct {
	client *common.Client
}

func NewClient(optFns ...common.OptionsFunc) (*Client, error) {
	client, err := common.NewClient(endpoint, service, optFns...)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string, params any) (*resty.Request, error) {
	return c.client.NewRequest(method, path, params)
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	return c.client.DoRequest(req)
}

func (c *Client) doRequestWithResult(req *resty.Request, res sdkResponse) (*resty.Response, error) {
	resp, err := c.client.DoRequestWithResult(req, res)
	if err == nil {
		if err := res.GetAPIError(); err != nil {
			return resp, fmt.Errorf("sdkerr: api error: %s", err.Error())
		}
	}

	return resp, err
}
