package dynv6

import (
	"context"
	"fmt"
	"net/http"
)

type AddRecordRequest struct {
	Type     *string `json:"type,omitempty"`
	Name     *string `json:"name,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Weight   *int    `json:"weight,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Data     *string `json:"data,omitempty"`
	Flags    *int    `json:"flags,omitempty"`
	Tag      *string `json:"tag,omitempty"`
}

type AddRecordResponse DNSRecord

func (c *Client) AddRecord(zoneID int64, req *AddRecordRequest) (*AddRecordResponse, error) {
	return c.AddRecordWithContext(context.Background(), zoneID, req)
}

func (c *Client) AddRecordWithContext(ctx context.Context, zoneID int64, req *AddRecordRequest) (*AddRecordResponse, error) {
	if zoneID == 0 {
		return nil, fmt.Errorf("sdkerr: unset zoneID")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/zones/%d/records", zoneID))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &AddRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
