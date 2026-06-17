package oss

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	bucket string

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
	if opts.AccessKeySecret == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	if endpoint == "" {
		if opts.Bucket == "" {
			endpoint = fmt.Sprintf("https://oss-%s.aliyuncs.com", url.PathEscape(opts.Region))
		} else {
			endpoint = fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", url.PathEscape(opts.Bucket), url.PathEscape(opts.Region))
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
			// https://help.aliyun.com/zh/oss/developer-reference/recommend-to-use-signature-version-4
			// https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4

			method := strings.ToUpper(req.Method)

			nowUtc := time.Now().UTC()
			headerDateStr := nowUtc.Format(http.TimeFormat)
			requestDateStr := nowUtc.Format("20060102T150405Z")
			signDateStr := nowUtc.Format("20060102")

			requestResStr := req.Header.Get("X-API-Resource")
			req.Header.Del("X-API-Resource")

			canonicalUrl := escapePath(req.URL.Path)
			if canonicalUrl == "" {
				canonicalUrl = "/"
			}
			if canonicalUrl == "/" && requestResStr != "" {
				canonicalUrl = escapePath(requestResStr)
			}

			canonicalQueryStr := ""
			if len(req.URL.Query()) > 0 {
				query := req.URL.Query()

				keys := make([]string, 0, len(query))
				for key := range query {
					keys = append(keys, key)
				}
				sort.Strings(keys)

				for i, key := range keys {
					if i > 0 {
						canonicalQueryStr += "&"
					}

					value := query.Get(key)
					if value == "" {
						canonicalQueryStr += escapeQuery(key)
					} else {
						canonicalQueryStr += escapeQuery(key) + "=" + escapeQuery(value)
					}
				}
			}

			canonicalHeaders := ""
			additionalHeaders := ""
			if len(req.Header) > 0 {
				if req.Header.Get("X-OSS-Date") == "" {
					req.Header.Set("X-OSS-Date", requestDateStr)
				}
				if req.Header.Get("X-OSS-Content-SHA256") == "" {
					req.Header.Set("X-OSS-Content-SHA256", "UNSIGNED-PAYLOAD")
				}

				keys := make([]string, 0, len(req.Header))
				for key := range req.Header {
					key = strings.ToLower(key)
					if strings.HasPrefix(key, "x-oss-") {
						keys = append(keys, key)
					}
					if key == "content-type" || key == "content-md5" {
						keys = append(keys, key)
					}
				}
				sort.Strings(keys)

				for i, key := range keys {
					if i > 0 {
						canonicalHeaders += "\n"
					}

					value := strings.TrimSpace(req.Header.Get(key))
					canonicalHeaders += key + ":" + value
				}

				canonicalHeaders += "\n"
			}

			hashedPayload := req.Header.Get("X-OSS-Content-SHA256")

			canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, canonicalUrl, canonicalQueryStr, canonicalHeaders, additionalHeaders, hashedPayload)
			canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
			canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

			const signAlgorithmHeader = "OSS4-HMAC-SHA256"
			scope := fmt.Sprintf("%s/%s/oss/aliyun_v4_request", signDateStr, opts.Region)
			stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", signAlgorithmHeader, requestDateStr, scope, canonicalRequestHashHex)

			var h hash.Hash
			h = hmac.New(sha256.New, []byte("aliyun_v4"+opts.AccessKeySecret))
			h.Write([]byte(signDateStr))
			kDate := h.Sum(nil)
			h = hmac.New(sha256.New, kDate)
			h.Write([]byte(opts.Region))
			kRegion := h.Sum(nil)
			h = hmac.New(sha256.New, kRegion)
			h.Write([]byte("oss"))
			kService := h.Sum(nil)
			h = hmac.New(sha256.New, kService)
			h.Write([]byte("aliyun_v4_request"))
			kSigning := h.Sum(nil)

			h = hmac.New(sha256.New, kSigning)
			h.Write([]byte(stringToSign))
			signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

			req.Header.Set("Authorization", fmt.Sprintf("%s Credential=%s/%s, Signature=%s", signAlgorithmHeader, opts.AccessKeyId, scope, signature))
			req.Header.Set("Date", headerDateStr)

			return nil
		})

	return &Client{
		bucket: opts.Bucket,
		rc:     restyClient,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string, resource string, params any) (*resty.Request, error) {
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
	req.SetHeader("X-API-Resource", resource)
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

func escapeQuery(str string) string {
	res := url.QueryEscape(str)
	res = strings.ReplaceAll(res, "+", "%20")
	return res
}

func escapePath(path string) string {
	var buf bytes.Buffer
	for i := 0; i < len(path); i++ {
		c := path[i]
		noEscape := (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' ||
			c == '.' ||
			c == '_' ||
			c == '~' ||
			c == '/'
		if noEscape {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}
