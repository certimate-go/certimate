package lvdn

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type QueryCertDetailRequest struct {
	Id   *int64  `json:"id,omitempty"   url:"id,omitempty"`
	Name *string `json:"name,omitempty" url:"name,omitempty"`
}

type QueryCertDetailResponse struct {
	sdkResponseBase

	Result *CertDetail `json:"result,omitempty"`
}

func (c *Client) QueryCertDetail(req *QueryCertDetailRequest) (*QueryCertDetailResponse, error) {
	return c.QueryCertDetailWithContext(context.Background(), req)
}

func (c *Client) QueryCertDetailWithContext(ctx context.Context, req *QueryCertDetailRequest) (*QueryCertDetailResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/cert/query-cert-detail")
	if err != nil {
		return nil, err
	} else {
		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &QueryCertDetailResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
