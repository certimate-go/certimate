package dynv6

import (
	"context"
	"fmt"
	"net/http"
)

type ListRecordsResponse []*DNSRecord

func (c *Client) ListRecords(zoneID int64) (*ListRecordsResponse, error) {
	return c.ListRecordsWithContext(context.Background(), zoneID)
}

func (c *Client) ListRecordsWithContext(ctx context.Context, zoneID int64) (*ListRecordsResponse, error) {
	if zoneID == 0 {
		return nil, fmt.Errorf("sdkerr: unset zoneID")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/zones/%d/records", zoneID))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ListRecordsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
