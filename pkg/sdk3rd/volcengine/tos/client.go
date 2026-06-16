package tos

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
		if opts.Region == "" {
			endpoint = fmt.Sprintf("https://tos-%s.volces.com", url.PathEscape(opts.Region))
		} else {
			endpoint = fmt.Sprintf("https://%s.tos-%s.volces.com", url.PathEscape(opts.Bucket), url.PathEscape(opts.Region))
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
			// https://www.volcengine.com/docs/6349/74839
			// https://docs.byteplus.com/en/docs/tos/reference-signature-mechanism_1

			method := strings.ToUpper(req.Method)

			nowUtc := time.Now().UTC()
			headerDateStr := nowUtc.Format(http.TimeFormat)
			requestDateStr := nowUtc.Format("20060102T150405Z")
			credentialDateStr := nowUtc.Format("20060102")

			canonicalUrl := req.URL.Path
			if canonicalUrl == "" {
				canonicalUrl = "/"
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
					canonicalQueryStr += fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value))
				}
			}

			canonicalHeaders := ""
			signedHeaders := ""
			if len(req.Header) > 0 || req.Host != "" {
				if req.Header.Get("Host") == "" {
					req.Header.Set("Host", req.Host)
				}
				if req.Header.Get("X-TOS-Date") == "" {
					req.Header.Set("X-TOS-Date", requestDateStr)
				}

				keys := make([]string, 0, len(req.Header))
				for key := range req.Header {
					key = strings.ToLower(key)
					if strings.HasPrefix(key, "x-tos-") || key == "host" {
						keys = append(keys, key)
					}
					if key == "content-type" && req.Header.Get("X-TOS-Content-SHA256") != "" {
						keys = append(keys, key)
					}
				}
				sort.Strings(keys)

				for i, key := range keys {
					if i > 0 {
						canonicalHeaders += "\n"
						signedHeaders += ";"
					}

					value := strings.TrimSpace(req.Header.Get(key))
					canonicalHeaders += fmt.Sprintf("%s:%s", key, value)
					signedHeaders += key
				}

				canonicalHeaders += "\n"
			}

			payloadSha256Str := req.Header.Get("X-TOS-Content-SHA256")
			if payloadSha256Str == "" {
				payloadSha256 := sha256.Sum256([]byte{})
				payloadSha256Str = strings.ToLower(hex.EncodeToString(payloadSha256[:]))
			}

			canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, canonicalUrl, canonicalQueryStr, canonicalHeaders, signedHeaders, payloadSha256Str)
			canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
			canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

			const signAlgorithmHeader = "TOS4-HMAC-SHA256"
			credentialScope := fmt.Sprintf("%s/%s/tos/request", credentialDateStr, opts.Region)
			stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", signAlgorithmHeader, requestDateStr, credentialScope, canonicalRequestHashHex)

			var h hash.Hash
			h = hmac.New(sha256.New, []byte(opts.SecretAccessKey))
			h.Write([]byte(credentialDateStr))
			kDate := h.Sum(nil)
			h = hmac.New(sha256.New, kDate)
			h.Write([]byte(opts.Region))
			kRegion := h.Sum(nil)
			h = hmac.New(sha256.New, kRegion)
			h.Write([]byte("tos"))
			kService := h.Sum(nil)
			h = hmac.New(sha256.New, kService)
			h.Write([]byte("request"))
			kSigning := h.Sum(nil)

			h = hmac.New(sha256.New, kSigning)
			h.Write([]byte(stringToSign))
			signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

			req.Header.Set("Authorization", fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s", signAlgorithmHeader, opts.AccessKeyId, credentialScope, signedHeaders, signature))
			req.Header.Set("Date", headerDateStr)

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
		// 目前仅支持 JSON 请求体，仅适用于非 S3 兼容接口
		payloadb, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}

		payloadStr = string(payloadb)
		contentType = "application/json"
	}
	payloadMd5 := md5.Sum([]byte(payloadStr))
	payloadMd5Encoded := base64.StdEncoding.EncodeToString(payloadMd5[:])

	req := c.rc.R()
	req.Method = method
	req.URL = requestUrl.Path
	req.QueryParam = requestUrl.Query()
	req.SetBody(payloadStr)
	req.SetHeader("Content-MD5", payloadMd5Encoded)
	req.SetHeader("Content-Type", contentType)
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
