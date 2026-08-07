// A simple SDK client for aaPanel.
// API documentation: https://www.aapanel.com/docs/api/api-list.html
package btpanel

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	apiKey string

	rc *resty.Client
}

func NewClient(serverUrl string, optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if opts.ApiKey == "" {
		return nil, fmt.Errorf("sdkerr: unset apiKey")
	}

	httper := resty.New().
		SetBaseURL(strings.TrimSuffix(serverUrl, "/")).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("User-Agent", app.AppUserAgent)

	return &Client{
		apiKey: opts.ApiKey,
		rc:     httper,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.rc.SetTLSClientConfig(config)
	return c
}

func (c *Client) newRequest(method string, path string, params any) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	paramsMap := make(map[string]string)
	if params != nil {
		temp := make(map[string]any)
		jsonb, _ := json.Marshal(params)
		json.Unmarshal(jsonb, &temp)
		for k, v := range temp {
			if v == nil {
				continue
			}

			switch reflect.Indirect(reflect.ValueOf(v)).Kind() {
			case reflect.String:
				paramsMap[k] = v.(string)

			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				paramsMap[k] = fmt.Sprintf("%v", v)

			default:
				if t, ok := v.(time.Time); ok {
					paramsMap[k] = t.Format(time.RFC3339)
				} else {
					jsonb, _ := json.Marshal(v)
					paramsMap[k] = string(jsonb)
				}
			}
		}
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	paramsMap["request_time"] = timestamp
	paramsMap["request_token"] = generateSignature(timestamp, c.apiKey)

	req := c.rc.R()
	req.Method = method
	req.URL = path
	req.SetFormData(paramsMap)

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
			if rStatus := res.GetStatus(); rStatus != nil && !*rStatus {
				if res.GetMessage() == nil {
					return resp, fmt.Errorf("sdkerr: api error: unknown error")
				} else {
					return resp, fmt.Errorf("sdkerr: api error: message='%s'", *res.GetMessage())
				}
			}
		}
	}

	return resp, nil
}

func generateSignature(timestamp string, apiKey string) string {
	keyMd5 := md5.Sum([]byte(apiKey))
	keyMd5Hex := strings.ToLower(hex.EncodeToString(keyMd5[:]))

	signMd5 := md5.Sum([]byte(timestamp + keyMd5Hex))
	signMd5Hex := strings.ToLower(hex.EncodeToString(signMd5[:]))

	return signMd5Hex
}
