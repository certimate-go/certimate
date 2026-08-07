package linode

import (
	"context"
	"fmt"
	"net/http"
)

type DeleteDomainRecordResponse struct {
	sdkResponseBase
}

func (c *Client) DeleteDomainRecord(domainId int, recordId int) (*DeleteDomainRecordResponse, error) {
	return c.DeleteDomainRecordWithContext(context.Background(), domainId, recordId)
}

func (c *Client) DeleteDomainRecordWithContext(ctx context.Context, domainId int, recordId int) (*DeleteDomainRecordResponse, error) {
	if domainId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainId")
	}
	if recordId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset recordId")
	}

	path := fmt.Sprintf("/domains/%d/records/%d", domainId, recordId)
	httpreq, err := c.newRequest(http.MethodDelete, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DeleteDomainRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
