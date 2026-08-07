package console

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type GetBucketsRequest struct {
	BucketName    string `json:"status"        url:"bucket_name"`
	BusinessType  string `json:"business_type" url:"business_type"`
	Type          string `json:"type"          url:"type"`
	Status        string `json:"state"         url:"state"`
	Tag           string `json:"tag"           url:"tag"`
	IsSecurityCDN bool   `json:"security_cdn"  url:"security_cdn"`
	WithDomains   bool   `json:"with_domains"  url:"with_domains"`
	Page          int32  `json:"page"          url:"page"`
	PerPage       int32  `json:"perPage"       url:"perPage"`
}

type GetBucketsResponse struct {
	sdkResponseBase
	Data *struct {
		sdkResponseBaseData
		Buckets []*BucketInfo `json:"buckets"`
		Pager   BucketPager   `json:"pager"`
	} `json:"data,omitempty"`
}

type BucketInfo struct {
	BucketName    string          `json:"bucket_name"`
	BusinessType  string          `json:"business_type"`
	Type          string          `json:"type"`
	Status        string          `json:"status"`
	Tag           string          `json:"tag"`
	IsFusionCDN   bool            `json:"fusion_cdn"`
	IsSecurityCDN bool            `json:"security_cdn"`
	Domains       []*BucketDomain `json:"domains"`
	Visible       bool            `json:"visible"`
	CreatedAt     string          `json:"created_at"`
}

type BucketDomain struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
}

type BucketPager struct {
	Page  int32 `json:"page"`
	Pages int64 `json:"pages"`
	Total int64 `json:"total"`
}

func (c *Client) GetBuckets(req *GetBucketsRequest) (*GetBucketsResponse, error) {
	return c.GetBucketsWithContext(context.Background(), req)
}

func (c *Client) GetBucketsWithContext(ctx context.Context, req *GetBucketsRequest) (*GetBucketsResponse, error) {
	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/api/v2/buckets")
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

	result := &GetBucketsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
