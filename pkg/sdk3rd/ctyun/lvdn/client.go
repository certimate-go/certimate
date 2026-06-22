// A simple SDK client for StateCloud LVDN.
// API documentation: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=125&vid=261
package lvdn

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	common "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/zz-shared-common"
)

const endpoint = "https://ctlvdn-global.ctapi.ctyun.cn"

type Client struct {
	client *common.Client
}

func NewClient(optFns ...common.OptionsFunc) (*Client, error) {
	client, err := common.NewClient(endpoint, optFns...)
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
