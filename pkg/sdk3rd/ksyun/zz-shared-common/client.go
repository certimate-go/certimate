package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	rc *resty.Client
}

func NewClient(baseUrl string, service string, optFns ...OptionsFunc) (*Client, error) {
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if baseUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset baseUrl")
	}
	if _, err := url.Parse(baseUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid baseUrl: %w", err)
	}
	if service == "" {
		return nil, fmt.Errorf("sdkerr: unset service")
	}
	if options.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if options.SecretAccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	signer := &signer{
		accessKeyId:     options.AccessKeyId,
		secretAccessKey: options.SecretAccessKey,
		service:         service,
	}
	httper := resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if err := signer.Sign(req); err != nil {
				return fmt.Errorf("sdkerr: sign error: %w", err)
			}

			return nil
		})

	return &Client{rc: httper}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) NewRequest(method string, path string, params any) (*resty.Request, error) {
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
				jsonb, _ := json.Marshal(v)
				paramsMap[k] = string(jsonb)
			}
		}
	}

	req := c.rc.R()
	req.Method = method
	req.URL = path
	if strings.ToUpper(method) == http.MethodGet {
		req.SetQueryParams(paramsMap)
	} else {
		req.SetBody(paramsMap)
	}

	return req, nil
}

func (c *Client) DoRequest(req *resty.Request) (*resty.Response, error) {
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

func (c *Client) DoRequestWithResult(req *resty.Request, res any) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.DoRequest(req)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &res)
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}
