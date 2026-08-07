package cms

import (
	"context"
	"net/http"
)

type GetCertificateListRequest struct {
	Status   *string `json:"status,omitempty"`
	Keyword  *string `json:"keyword,omitempty"`
	PageNum  *int32  `json:"pageNum,omitempty"`
	PageSize *int32  `json:"pageSize,omitempty"`
	Origin   *string `json:"origin,omitempty"`
}

type GetCertificateListResponse struct {
	sdkResponseBase

	ReturnObj *struct {
		List      []*Certificate `json:"list,omitempty"`
		TotalSize int32          `json:"totalSize,omitempty"`
	} `json:"returnObj,omitempty"`
}

func (c *Client) GetCertificateList(req *GetCertificateListRequest) (*GetCertificateListResponse, error) {
	return c.GetCertificateListWithContext(context.Background(), req)
}

func (c *Client) GetCertificateListWithContext(ctx context.Context, req *GetCertificateListRequest) (*GetCertificateListResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/v1/certificate/list")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &GetCertificateListResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
