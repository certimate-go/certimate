package mohua

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ListVirtualHostDomainsResponse struct {
	sdkResponseBase

	Data []*DomainInfo `json:"data"`
}

func (c *Client) ListVirtualHostDomains(hostId string) (*ListVirtualHostDomainsResponse, error) {
	return c.ListVirtualHostDomainsWithContext(context.Background(), hostId)
}

func (c *Client) ListVirtualHostDomainsWithContext(ctx context.Context, hostId string) (*ListVirtualHostDomainsResponse, error) {
	if hostId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset hostId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/provision/custom/%s/domains", url.PathEscape(hostId))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(map[string]any{
			"func": "ListDomain",
		})
		httpreq.SetContext(ctx)
	}

	result := &ListVirtualHostDomainsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
