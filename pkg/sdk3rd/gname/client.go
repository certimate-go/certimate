package gname

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	appId  string
	appKey string

	client *resty.Client
}

func NewClient(appId, appKey string) (*Client, error) {
	if appId == "" {
		return nil, fmt.Errorf("sdkerr: unset appId")
	}
	if appKey == "" {
		return nil, fmt.Errorf("sdkerr: unset appKey")
	}

	client := resty.New().
		SetBaseURL("https://api.gname.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("User-Agent", app.AppUserAgent)

	return &Client{
		appId:  appId,
		appKey: appKey,
		client: client,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string, params any) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	data := make(map[string]string)
	if params != nil {
		temp := make(map[string]any)
		jsonb, _ := json.Marshal(params)
		json.Unmarshal(jsonb, &temp)
		for k, v := range temp {
			if v == nil {
				continue
			}

			data[k] = fmt.Sprintf("%v", v)
		}
	}

	data["appid"] = c.appId
	data["gntime"] = fmt.Sprintf("%d", time.Now().Unix())
	data["gntoken"] = generateSignature(data, c.appKey)

	req := c.client.R()
	req.Method = method
	req.URL = path
	req.SetFormData(data)
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	// WARN:
	//   PLEASE DO NOT USE `req.SetBody` or `req.SetFormData` HERE! USE `newRequest` INSTEAD.
	//   PLEASE DO NOT USE `req.SetResult` or `req.SetError` HERE! USE `doRequestWithResult` INSTEAD.

	resp, err := req.Send()
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
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
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		} else {
			if tcode := res.GetCode(); tcode != 1 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', message='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}

func generateSignature(params map[string]string, appKey string) string {
	// Step 1: Sort parameters by ASCII order
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Step 2: Create string A with URL-encoded values
	var pairs []string
	for _, k := range keys {
		encodedValue := url.QueryEscape(params[k])
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, encodedValue))
	}
	stringA := strings.Join(pairs, "&")

	// Step 3: Append appkey to create string B
	stringB := stringA + appKey

	// Step 4: Calculate MD5 and convert to uppercase
	hash := md5.Sum([]byte(stringB))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}
