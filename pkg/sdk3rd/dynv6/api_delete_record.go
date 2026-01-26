package dynv6

import (
	"context"
	"fmt"
	"net/http"
)

type DeleteRecordResponse DNSRecord

func (c *Client) DeleteRecord(zoneID int64, recordID int64) (*DeleteRecordResponse, error) {
	return c.DeleteRecordWithContext(context.Background(), zoneID, recordID)
}

func (c *Client) DeleteRecordWithContext(ctx context.Context, zoneID int64, recordID int64) (*DeleteRecordResponse, error) {
	if zoneID == 0 {
		return nil, fmt.Errorf("sdkerr: unset zoneID")
	}
	if recordID == 0 {
		return nil, fmt.Errorf("sdkerr: unset recordID")
	}

	httpreq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/zones/%d/records/%d", zoneID, recordID))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DeleteRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
