// A simple SDK client for HuaweiCloud OBS.
// API documentation: https://support.huaweicloud.com/api-obs/obs_04_0005.html
package obs

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
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

func NewClient(endpoint string, optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.Region == "" {
		return nil, fmt.Errorf("sdkerr: unset region")
	}
	if opts.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if opts.SecretAccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	if endpoint == "" {
		if opts.Bucket == "" {
			endpoint = fmt.Sprintf("https://obs.%s.myhuaweicloud.com", url.PathEscape(opts.Region))
		} else {
			endpoint = fmt.Sprintf("https://%s.obs.%s.myhuaweicloud.com", url.PathEscape(opts.Bucket), url.PathEscape(opts.Region))
		}
	} else {
		if baseUrl, err := url.Parse(endpoint); err != nil {
			return nil, fmt.Errorf("sdkerr: invalid endpoint: %w", err)
		} else if baseUrl.Scheme == "" {
			endpoint = "https://" + endpoint
		}
	}

	signer := &signer{
		accessKeyId:     opts.AccessKeyId,
		secretAccessKey: opts.SecretAccessKey,
		bucket:          opts.Bucket,
	}
	httper := resty.New().
		SetBaseURL(endpoint).
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

func (c *Client) newRequest(method string, path string, params any) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	requestUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("sdkerr: invalid path: %w", err)
	} else if requestUrl.IsAbs() {
		return nil, fmt.Errorf("sdkerr: path should be relative")
	}

	payloadStr := ""
	contentType := ""
	if params != nil {
		// 目前仅支持 XML 请求体，仅适用于非 S3 兼容接口
		payloadb, err := xml.Marshal(params)
		if err != nil {
			return nil, err
		}

		payloadStr = string(payloadb)
		contentType = "application/xml"
	}
	payloadMd5 := md5.Sum([]byte(payloadStr))
	payloadMd5Encoded := base64.StdEncoding.EncodeToString(payloadMd5[:])

	req := c.rc.R()
	req.Method = method
	req.URL = requestUrl.Path
	req.QueryParam = requestUrl.Query()
	req.SetHeader("Content-MD5", payloadMd5Encoded)
	req.SetHeader("Content-Type", contentType)
	req.SetBody(payloadStr)

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
			xml.Unmarshal(resp.Body(), &res)
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		if err := xml.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}
