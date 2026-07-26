package lvdn

import (
	"context"
	"net/http"
)

type QueryDomainDetailRequest struct {
	Domain      *string `json:"domain,omitempty"       url:"domain,omitempty"`
	ProductCode *string `json:"product_code,omitempty" url:"product_code,omitempty"`
}

type QueryDomainDetailResponse struct {
	sdkResponseBase

	ReturnObj *DomainDetail `json:"returnObj,omitempty"`
}

func (c *Client) QueryDomainDetail(req *QueryDomainDetailRequest) (*QueryDomainDetailResponse, error) {
	return c.QueryDomainDetailWithContext(context.Background(), req)
}

func (c *Client) QueryDomainDetailWithContext(ctx context.Context, req *QueryDomainDetailRequest) (*QueryDomainDetailResponse, error) {
	httpreq, err := c.domainClient.NewRequest(http.MethodPost, "/live/domain/query-domain-detail")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &QueryDomainDetailResponse{}
	if _, err := c.doDomainRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
