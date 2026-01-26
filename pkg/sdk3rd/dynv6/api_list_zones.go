package dynv6

import (
	"context"
	"net/http"
)

type ListZonesResponse []*ZoneRecord

func (c *Client) ListZones() (*ListZonesResponse, error) {
	return c.ListZonesWithContext(context.Background())
}

func (c *Client) ListZonesWithContext(ctx context.Context) (*ListZonesResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/zones")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ListZonesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
