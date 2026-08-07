package linode

import (
	"context"
	"fmt"
	"net/http"
)

type CreateDomainRecordRequest DomainRecord

type CreateDomainRecordResponse struct {
	sdkResponseBase

	DomainRecord `json:",inline"`
}

func (c *Client) CreateDomainRecord(domainId int, req *CreateDomainRecordRequest) (*CreateDomainRecordResponse, error) {
	return c.CreateDomainRecordWithContext(context.Background(), domainId, req)
}

func (c *Client) CreateDomainRecordWithContext(ctx context.Context, domainId int, req *CreateDomainRecordRequest) (*CreateDomainRecordResponse, error) {
	if domainId == 0 {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainId")
	}

	path := fmt.Sprintf("/domains/%d/records", domainId)
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CreateDomainRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
