package nginxproxymanager

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type NginxListStreamsRequest struct {
	Expand *string `json:"expand,omitempty" url:"expand,omitempty"`
}

type NginxListStreamsResponse = []*StreamHostRecord

func (c *Client) NginxListStreams(req *NginxListStreamsRequest) (*NginxListStreamsResponse, error) {
	return c.NginxListStreamsWithContext(context.Background(), req)
}

func (c *Client) NginxListStreamsWithContext(ctx context.Context, req *NginxListStreamsRequest) (*NginxListStreamsResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/nginx/streams")
	if err != nil {
		return nil, err
	} else {
		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &NginxListStreamsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
