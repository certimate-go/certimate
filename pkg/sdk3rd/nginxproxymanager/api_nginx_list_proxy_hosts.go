package nginxproxymanager

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type NginxListProxyHostsRequest struct {
	Expand *string `json:"expand,omitempty" url:"expand,omitempty"`
}

type NginxListProxyHostsResponse = []*ProxyHostRecord

func (c *Client) NginxListProxyHosts(req *NginxListProxyHostsRequest) (*NginxListProxyHostsResponse, error) {
	return c.NginxListProxyHostsWithContext(context.Background(), req)
}

func (c *Client) NginxListProxyHostsWithContext(ctx context.Context, req *NginxListProxyHostsRequest) (*NginxListProxyHostsResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/nginx/proxy-hosts")
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

	result := &NginxListProxyHostsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
