// A simple SDK client for AWS Elastic Load Balancing.
// API documentation: https://docs.aws.amazon.com/elasticloadbalancing/
package elasticloadbalancing

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	common "github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-common"
	smithyprotocolquery "github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/query"
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

	service := "elasticloadbalancing"
	region := strings.TrimSpace(options.Region)
	baseUrl, err := common.ResolveBaseEndpoint(service, region, common.EndpointVariantNone)
	if err != nil {
		return nil, fmt.Errorf("sdkerr: %w", err)
	}

	signer := common.NewSigner(options.AccessKeyId, options.SecretAccessKey, service, region)
	httper := resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Accept", "application/xml").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
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

func (c *Client) newRequest(params any) (*resty.Request, error) {
	req := c.rc.R()
	req.Method = http.MethodPost
	req.URL = "/"

	if params != nil {
		sezer := smithyprotocolquery.NewSerializer()
		sezer.UseOmitEmptyValue()
		paramsMap, _ := sezer.SerializeToMap(params)

		if paramsMap["Action"] == "" {
			return nil, fmt.Errorf("sdkerr: bad request: unset action in params")
		}
		if paramsMap["Version"] == "" {
			return nil, fmt.Errorf("sdkerr: bad request: unset version in params")
		}

		req.SetFormData(paramsMap)
	}

	// WARN:
	//   DO NOT CALL `req.SetBody` or `req.SetFormData` AGAIN! USE `newRequest` INSTEAD.
	//   DO NOT CALL `req.SetResult` or `req.SetError` LATER! USE `doRequestWithResult` INSTEAD.
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

func (c *Client) doRequestWithResult(req *resty.Request, res any) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.doRequest(req)
	if err != nil {
		if resp != nil {
			if sdkErr, _ := smithyprotocolquery.GetAPIErrorWithRawResponse(resp.Body(), resp.RawResponse); sdkErr != nil {
				return resp, sdkErr
			}
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		action := ""
		if req.FormData != nil {
			action = req.FormData.Get("Action")
		}

		dezer := smithyprotocolquery.NewDeserializer(action)
		if err := dezer.Deserialize(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}
