package cdnpro

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type GetHostnameDetailResponse struct {
	sdkResponseBase

	Hostname             string            `json:"hostname"`
	PropertyInProduction *HostnameProperty `json:"propertyInProduction,omitempty"`
	PropertyInStaging    *HostnameProperty `json:"propertyInStaging,omitempty"`
}

func (c *Client) GetHostnameDetail(hostname string) (*GetHostnameDetailResponse, error) {
	return c.GetHostnameDetailWithContext(context.Background(), hostname)
}

func (c *Client) GetHostnameDetailWithContext(ctx context.Context, hostname string) (*GetHostnameDetailResponse, error) {
	if hostname == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset hostname")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/cdn/hostnames/%s", url.PathEscape(hostname)))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetHostnameDetailResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
