package nginxproxymanager

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type NginxListRedirectionHostsRequest struct {
	Expand *string `json:"expand,omitempty" url:"expand,omitempty"`
}

type NginxListRedirectionHostsResponse = []*RedirectionHostRecord

func (c *Client) NginxListRedirectionHosts(req *NginxListRedirectionHostsRequest) (*NginxListRedirectionHostsResponse, error) {
	return c.NginxListRedirectionHostsWithContext(context.Background(), req)
}

func (c *Client) NginxListRedirectionHostsWithContext(ctx context.Context, req *NginxListRedirectionHostsRequest) (*NginxListRedirectionHostsResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/nginx/redirection-hosts")
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

	result := &NginxListRedirectionHostsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
