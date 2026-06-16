package tos

import (
	"context"
	"net/http"
)

type PutBucketCustomDomainRequest struct {
	CustomDomainRule *PutBucketCustomDomainRequestCustomDomainRule `json:",omitempty"`
}

type PutBucketCustomDomainRequestCustomDomainRule struct {
	CertId   string `json:",omitempty"`
	Domain   string `json:",omitempty"`
	Protocol string `json:",omitempty"`
}

type PutBucketCustomDomainResponse struct {
	sdkResponseBase
}

func (c *Client) PutBucketCustomDomain(req *PutBucketCustomDomainRequest) (*PutBucketCustomDomainResponse, error) {
	return c.PutBucketCustomDomainWithContext(context.Background(), req)
}

func (c *Client) PutBucketCustomDomainWithContext(ctx context.Context, req *PutBucketCustomDomainRequest) (*PutBucketCustomDomainResponse, error) {
	httpreq, err := c.newRequest(http.MethodPut, "/?customdomain", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &PutBucketCustomDomainResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
