package eomakers

import (
	"context"
	"fmt"
)

func NewDescribePagesZoneCustomDomainsReq() *DescribePagesZoneCustomDomainsReq {
	return &DescribePagesZoneCustomDomainsReq{Action: "DescribePagesZoneCustomDomains"}
}

func NewDescribeMakersZoneCustomDomainsReq() *DescribeMakersZoneCustomDomainsReq {
	return NewDescribePagesZoneCustomDomainsReq()
}

// DescribePagesZoneCustomDomains provides a method to
// describe zone custom domains with pages.
func (c Client) DescribePagesZoneCustomDomains(ctx context.Context, request *DescribePagesZoneCustomDomainsReq) (
	*DescribePagesZoneCustomDomainsResp, error,
) {
	if request == nil {
		request = NewDescribePagesZoneCustomDomainsReq()
	}

	var resp DescribePagesZoneCustomDomainsResp
	if err := doRequest(ctx, c.apiToken, request, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("sdkErr: Code %d: %v", resp.Code, resp.Data)
	}

	return &resp, nil
}

// DescribeMakersZoneCustomDomains is the alias of
// c.DescribePagesZoneCustomDomains due to the production rename
// reserved for future API migration
func (c Client) DescribeMakersZoneCustomDomains(ctx context.Context, request *DescribePagesZoneCustomDomainsReq) (
	*DescribeMakersZoneCustomDomainsResp, error,
) {
	if request == nil {
		request = NewDescribeMakersZoneCustomDomainsReq()
	}
	return c.DescribePagesZoneCustomDomains(ctx, request)
}
