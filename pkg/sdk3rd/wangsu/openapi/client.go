package openapi

import (
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
	accessKey string
	secretKey string

	client *resty.Client
}

func NewClient(accessKey, secretKey string) (*Client, error) {
	if accessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKey")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretKey")
	}

	client := resty.New().
		SetBaseURL("https://open.chinanetcenter.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("Host", "open.chinanetcenter.com").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// Step 1: Get request method
			method := req.Method
			method = strings.ToUpper(method)

			// Step 2: Get request path
			path := "/"
			if req.URL != nil {
				path = req.URL.Path
			}

			// Step 3: Get unencoded query string
			queryString := ""
			if method != http.MethodPost && req.URL != nil {
				queryString = req.URL.RawQuery

				s, err := url.QueryUnescape(queryString)
				if err != nil {
					return err
				}

				queryString = s
			}

			// Step 4: Get canonical headers & signed headers
			canonicalHeaders := "" +
				"content-type:" + strings.TrimSpace(strings.ToLower(req.Header.Get("Content-Type"))) + "\n" +
				"host:" + strings.TrimSpace(strings.ToLower(req.Header.Get("Host"))) + "\n"
			signedHeaders := "content-type;host"

			// Step 5: Get request payload
			payload := ""
			if method != http.MethodGet && req.Body != nil {
				reader, err := req.GetBody()
				if err != nil {
					return err
				}

				defer reader.Close()

				payloadb, err := io.ReadAll(reader)
				if err != nil {
					return err
				}

				payload = string(payloadb)
			}
			hashedPayload := sha256.Sum256([]byte(payload))
			hashedPayloadHex := strings.ToLower(hex.EncodeToString(hashedPayload[:]))

			// Step 6: Get timestamp
			var reqtime time.Time
			timestampString := req.Header.Get("X-CNC-Timestamp")
			if timestampString == "" {
				reqtime = time.Now().UTC()
				timestampString = fmt.Sprintf("%d", reqtime.Unix())
			} else {
				timestamp, err := strconv.ParseInt(timestampString, 10, 64)
				if err != nil {
					return err
				}
				reqtime = time.Unix(timestamp, 0).UTC()
			}

			// Step 7: Get canonical request string
			canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, path, queryString, canonicalHeaders, signedHeaders, hashedPayloadHex)
			hashedCanonicalRequest := sha256.Sum256([]byte(canonicalRequest))
			hashedCanonicalRequestHex := strings.ToLower(hex.EncodeToString(hashedCanonicalRequest[:]))

			// Step 8: String to sign
			const SignAlgorithmHeader = "CNC-HMAC-SHA256"
			stringToSign := fmt.Sprintf("%s\n%s\n%s", SignAlgorithmHeader, timestampString, hashedCanonicalRequestHex)
			hmac := hmac.New(sha256.New, []byte(secretKey))
			hmac.Write([]byte(stringToSign))
			sign := hmac.Sum(nil)
			signHex := strings.ToLower(hex.EncodeToString(sign))

			// Step 9: Add headers to request
			req.Header.Set("X-CNC-AccessKey", accessKey)
			req.Header.Set("X-CNC-Timestamp", timestampString)
			req.Header.Set("X-CNC-Auth-Method", "AKSK")
			req.Header.Set("Authorization", fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s", SignAlgorithmHeader, accessKey, signedHeaders, signHex))
			req.Header.Set("Date", reqtime.Format("Mon, 02 Jan 2006 15:04:05 GMT"))

			return nil
		})

	return &Client{
		accessKey: accessKey,
		secretKey: secretKey,
		client:    client,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) NewRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.client.R()
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
