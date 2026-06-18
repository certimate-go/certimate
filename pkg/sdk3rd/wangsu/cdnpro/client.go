// A simple SDK client for WangsuCloud CDN.
// API documentation: https://www.wangsu.com/document/api-doc/product
package cdnpro

import (
	"time"

	"github.com/go-resty/resty/v2"

	common "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/zz-shared-common"
)

type Client struct {
	client *common.Client
}

func NewClient(optFns ...common.OptionsFunc) (*Client, error) {
	client, err := common.NewClient(optFns...)
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
	return c.client.DoRequestWithResult(req, res)
}
