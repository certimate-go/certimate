package obs

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
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

	restyClient := resty.New().
		SetBaseURL(endpoint).
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// API 签名机制：
			// https://support.huaweicloud.com/api-obs/obs_04_0010.html

			canonicalizedHeaders := ""
			if len(req.Header) > 0 {
				keys := make([]string, 0, len(req.Header))
				for key := range req.Header {
					key = strings.ToLower(key)
					if strings.HasPrefix(key, "x-obs-") {
						keys = append(keys, key)
					}
				}
				sort.Strings(keys)

				for _, key := range keys {
					value := strings.TrimSpace(req.Header.Get(key))
					canonicalizedHeaders += fmt.Sprintf("%s:%s", key, url.QueryEscape(value))
					canonicalizedHeaders += "\n"
				}
			}

			bucketName := opts.Bucket
			objectName := strings.Trim(req.URL.Path, "/")
			canonicalizedResource := fmt.Sprintf("/%s/%s", bucketName, objectName)
			if bucketName == "" && objectName == "" {
				canonicalizedResource = "/"
			}
			if len(req.URL.Query()) > 0 {
				query := req.URL.Query()

				keys := make([]string, 0, len(query))
				for key := range query {
					keys = append(keys, key)
				}
				sort.Strings(keys)

				for i, key := range keys {
					if i == 0 {
						canonicalizedResource += "?"
					} else {
						canonicalizedResource += "&"
					}

					value := query.Get(key)
					if value == "" {
						canonicalizedResource += key
					} else {
						canonicalizedResource += fmt.Sprintf("%s=%s", strings.ToLower(key), url.QueryEscape(value))
					}
				}
			}

			method := strings.ToUpper(req.Method)

			dateStr := time.Now().UTC().Format(http.TimeFormat)

			contentMd5 := req.Header.Get("Content-MD5")
			contentType := req.Header.Get("Content-Type")

			stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s%s", method, contentMd5, contentType, dateStr, canonicalizedHeaders, canonicalizedResource)
			println("stringToSign:", stringToSign)

			h := hmac.New(sha1.New, []byte(opts.SecretAccessKey))
			h.Write([]byte(stringToSign))
			signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

			req.Header.Set("Authorization", fmt.Sprintf("OBS %s:%s", opts.AccessKeyId, signature))
			req.Header.Set("Date", dateStr)

			return nil
		})

	return &Client{rc: restyClient}, nil
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
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	// WARN:
	//   PLEASE DO NOT USE `req.SetBody` HERE! USE `newRequest` INSTEAD.
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
