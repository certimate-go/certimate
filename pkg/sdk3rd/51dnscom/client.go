package dnscom

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	apiKey    string
	apiSecret string

	client *resty.Client
}

func NewClient(apiKey, apiSecret string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("sdkerr: unset apiKey")
	}
	if apiSecret == "" {
		return nil, fmt.Errorf("sdkerr: unset apiSecret")
	}

	client := resty.New().
		SetBaseURL("https://www.51dns.com/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate")

	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    client,
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

	data["apiKey"] = c.apiKey
	data["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	data["hash"] = generateHash(data, c.apiSecret)

	req := c.client.R()
	req.Method = method
	req.URL = path
	req.SetBody(data)
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
			if tcode := res.GetCode(); tcode != 0 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', message='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}

func generateHash(params map[string]string, secert string) string {
	var keyList []string
	for k := range params {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	var hashString string
	for _, key := range keyList {
		if hashString == "" {
			hashString += key + "=" + params[key]
		} else {
			hashString += "&" + key + "=" + params[key]
		}
	}

	m := md5.New()
	m.Write([]byte(hashString + secert))
	cipherStr := m.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
