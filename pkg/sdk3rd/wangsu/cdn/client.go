package cdn

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/openapi"
)

type Client struct {
	client *openapi.Client
}

func NewClient(accessKey, secretKey string) (*Client, error) {
	client, err := openapi.NewClient(accessKey, secretKey)
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

func (c *Client) doRequestWithResult(req *resty.Request, res apiResponse) (*resty.Response, error) {
	resp, err := c.client.DoRequestWithResult(req, res)
	if err == nil {
		if tcode := res.GetCode(); tcode != "" && tcode != "0" {
			return resp, fmt.Errorf("sdkerr: api error: code='%s', message='%s'", tcode, res.GetMessage())
		}
	}

	return resp, err
}
