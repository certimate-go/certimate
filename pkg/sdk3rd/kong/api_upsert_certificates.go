package kong

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpsertCertificateRequest Certificate

type UpsertCertificateResponse Certificate

func (c *Client) UpsertCertificate(certificateId string, req *UpsertCertificateRequest) (*UpsertCertificateResponse, error) {
	return c.UpsertCertificateWithContext(context.Background(), certificateId, req)
}

func (c *Client) UpsertCertificateWithContext(ctx context.Context, certificateId string, req *UpsertCertificateRequest) (*UpsertCertificateResponse, error) {
	if certificateId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset certificateId")
	}

	httpreq, err := c.newRequest(http.MethodPut, fmt.Sprintf("/certificates/%s", url.PathEscape(certificateId)))
	if err != nil {
		return nil, err
	} else {
		req.Id = &certificateId

		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpsertCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
