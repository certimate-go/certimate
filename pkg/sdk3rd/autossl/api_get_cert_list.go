package autossl

import (
	"context"
	"net/http"
	"strconv"
)

const (
	certListDefaultPageSize = 50
	certListMaxPageSize     = 50
)

type GetCertListRequest struct {
	Page int
	Size int
}

type GetCertListResponse struct {
	sdkResponseBase

	Items []*CertificateInfo `json:"-"`
}

func (r *GetCertListResponse) GetItems() []*CertificateInfo {
	return r.Items
}

// 获取证书列表。
//
// REF: https://www.auto-ssl.cn/docs/OpenAPI.html#getcertlist
func (c *Client) GetCertList(req *GetCertListRequest) (*GetCertListResponse, error) {
	return c.GetCertListWithContext(context.Background(), req)
}

func (c *Client) GetCertListWithContext(ctx context.Context, req *GetCertListRequest) (*GetCertListResponse, error) {
	if req == nil {
		req = &GetCertListRequest{}
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 || size > certListMaxPageSize {
		size = certListDefaultPageSize
	}

	httpreq, err := c.newRequest(http.MethodGet, "/v1/api/cert/getCertList")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParam("page", strconv.Itoa(page))
		httpreq.SetQueryParam("size", strconv.Itoa(size))
		httpreq.SetContext(ctx)
	}

	result := &GetCertListResponse{}
	resp, err := c.doRequestWithResult(httpreq, result)
	if err != nil {
		return result, err
	}

	items, err := parseCertificateItems(resp.Body())
	if err != nil {
		return result, err
	} else {
		result.Items = items
	}

	return result, nil
}

// 遍历全部分页获取证书列表。
func (c *Client) ListAllCertificatesWithContext(ctx context.Context) ([]*CertificateInfo, error) {
	page := 1
	allItems := make([]*CertificateInfo, 0)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.GetCertListWithContext(ctx, &GetCertListRequest{Page: page, Size: certListMaxPageSize})
		if err != nil {
			return nil, err
		}

		items := resp.GetItems()
		if len(items) == 0 {
			break
		}

		allItems = append(allItems, items...)
		if len(items) < certListMaxPageSize {
			break
		}

		page++
	}

	return allItems, nil
}
