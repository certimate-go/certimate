package dns

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	rc *resty.Client
}

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if options.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if options.SecretAccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	restyClient := resty.New().
		SetBaseURL("http://api.routewize.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Host", "api.routewize.com").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// 签名机制：
			// https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/dns/calling_method/signature/

			date := time.Now().UTC().Format(time.RFC1123)

			verb := req.Method

			canonicalizedResource := "/"
			if req.URL != nil {
				canonicalizedResource = req.URL.Path
				if req.URL.RawQuery != "" {
					values, _ := url.ParseQuery(req.URL.RawQuery)
					canonicalizedResource += "?" + values.Encode()
				}
			}

			stringToSign := verb + "\n" +
				date + "\n" +
				canonicalizedResource

			h := hmac.New(sha256.New, []byte(options.SecretAccessKey))
			h.Write([]byte(stringToSign))
			signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

			req.Header.Set("Date", date)
			req.Header.Set("Authorization", fmt.Sprintf("QC-HMAC-SHA256 %s:%s", options.AccessKeyId, signature))

			return nil
		})

	return &Client{rc: restyClient}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.rc.R()
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
			if rCode := res.GetCode(); rCode != 0 {
				return resp, fmt.Errorf("sdkerr: code='%d', message='%s'", rCode, res.GetMessage())
			}
		}
	}

	return resp, nil
}
