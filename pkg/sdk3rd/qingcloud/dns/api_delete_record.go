package dns

import (
	"context"
	"fmt"
	"net/http"
)

type DeleteRecordResponse struct {
	sdkResponseBase
}

func (c *Client) DeleteRecord(recordIds []*int64) (*DeleteRecordResponse, error) {
	return c.DeleteRecordWithContext(context.Background(), recordIds)
}

func (c *Client) DeleteRecordWithContext(ctx context.Context, recordIds []*int64) (*DeleteRecordResponse, error) {
	if len(recordIds) == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset recordIds")
	}

	httpreq, err := c.newRequest(http.MethodPost, "/v1/change_record_status/")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(map[string]any{
			"ids":    recordIds,
			"action": "delete",
			"target": "record",
		})
		httpreq.SetContext(ctx)
	}

	result := &DeleteRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
