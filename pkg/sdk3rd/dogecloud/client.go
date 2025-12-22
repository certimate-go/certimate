package dogecloud

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
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
		SetBaseURL("https://api.dogecloud.com").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate").
		SetPreRequestHook(func(ctx *resty.Client, req *http.Request) error {
			requestUrl := req.URL.Path
			requestQuery := req.URL.Query().Encode()
			if requestQuery != "" {
				requestUrl += "?" + requestQuery
			}

			payload := ""
			if req.Body != nil {
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

			stringToSign := fmt.Sprintf("%s\n%s", requestUrl, payload)
			mac := hmac.New(sha1.New, []byte(secretKey))
			mac.Write([]byte(stringToSign))
			sign := hex.EncodeToString(mac.Sum(nil))

			req.Header.Set("Authorization", fmt.Sprintf("TOKEN %s:%s", accessKey, sign))

			return nil
		})

	return &Client{client}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
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
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		} else {
			if tcode := res.GetCode(); tcode != 0 && tcode != 200 {
				return resp, fmt.Errorf("sdkerr: code='%d', msg='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}
