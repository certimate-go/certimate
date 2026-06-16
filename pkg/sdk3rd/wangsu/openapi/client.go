package openapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	if options.AccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKey")
	}
	if options.SecretKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretKey")
	}

	restyClient := resty.New().
		SetBaseURL("https://open.chinanetcenter.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// API 签名机制：
			// https://www.wangsu.com/document/openapi/api-authentication

			method := strings.ToUpper(req.Method)

			path := "/"
			if req.URL != nil {
				path = req.URL.Path
			}

			queryStr := ""
			if method != http.MethodPost && req.URL != nil {
				queryStr = req.URL.RawQuery

				s, err := url.QueryUnescape(queryStr)
				if err != nil {
					return err
				}

				queryStr = s
			}

			canonicalHeaders := "" +
				"content-type:" + strings.TrimSpace(strings.ToLower(req.Header.Get("Content-Type"))) + "\n" +
				"host:" + strings.TrimSpace(strings.ToLower(req.Host)) + "\n"
			signedHeaders := "content-type;host"

			payloadStr := ""
			if method != http.MethodGet && req.Body != nil {
				payloadb, err := io.ReadAll(req.Body)
				if err != nil {
					return err
				}

				payloadStr = string(payloadb)
				req.Body = io.NopCloser(bytes.NewReader(payloadb))
			}
			payloadHash := sha256.Sum256([]byte(payloadStr))
			payloadHashHex := strings.ToLower(hex.EncodeToString(payloadHash[:]))

			nowUtc := time.Now().UTC()
			timestampStr := req.Header.Get("X-CNC-Timestamp")
			if timestampStr == "" {
				timestampStr = fmt.Sprintf("%d", nowUtc.Unix())
			} else {
				timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
				if err != nil {
					return err
				}
				nowUtc = time.Unix(timestamp, 0).UTC()
			}

			canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, path, queryStr, canonicalHeaders, signedHeaders, payloadHashHex)
			canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
			canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

			const signAlgorithmHeader = "CNC-HMAC-SHA256"
			stringToSign := fmt.Sprintf("%s\n%s\n%s", signAlgorithmHeader, timestampStr, canonicalRequestHashHex)

			h := hmac.New(sha256.New, []byte(options.SecretKey))
			h.Write([]byte(stringToSign))
			signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

			req.Header.Set("Authorization", fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s", signAlgorithmHeader, options.AccessKey, signedHeaders, signature))
			req.Header.Set("Date", nowUtc.Format(http.TimeFormat))
			req.Header.Set("X-CNC-Auth-Method", "AKSK")
			req.Header.Set("X-CNC-AccessKey", options.AccessKey)
			req.Header.Set("X-CNC-Timestamp", timestampStr)

			return nil
		})

	return &Client{rc: restyClient}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) NewRequest(method string, path string) (*resty.Request, error) {
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

func (c *Client) DoRequest(req *resty.Request) (*resty.Response, error) {
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
