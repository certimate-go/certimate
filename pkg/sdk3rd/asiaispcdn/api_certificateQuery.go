package asiaispcdn

import (
	"context"
	"fmt"
	"net/http"
)

type CertificateQueryResponse struct {
	sdkResponseBase

	Data *CertificateDetail `json:"data,omitempty"`
}

func (c *Client) CertificateQuery(certId string) (*CertificateQueryResponse, error) {
	return c.CertificateQueryWithContext(context.Background(), certId)
}

func (c *Client) CertificateQueryWithContext(ctx context.Context, certId string) (*CertificateQueryResponse, error) {
	if certId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certId")
	}

	httpreq, err := c.newRequest(http.MethodGet, nil, "certificateQuery")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParam("certId", certId)
		httpreq.SetContext(ctx)
	}

	result := &CertificateQueryResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
