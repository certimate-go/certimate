package vod20260101

import (
	"github.com/volcengine/volcengine-go-sdk/service/vod20260101"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opListVodDomains = "ListVodDomains"

func (c *VOD20260101) ListVodDomainsRequest(input *ListVodDomainsInput) (req *request.Request, output *ListVodDomainsOutput) {
	op := &request.Operation{
		Name:       opListVodDomains,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListVodDomainsInput{}
	}

	output = &ListVodDomainsOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *VOD20260101) ListVodDomainsWithContext(ctx volcengine.Context, input *ListVodDomainsInput, opts ...request.Option) (*ListVodDomainsOutput, error) {
	req, out := c.ListVodDomainsRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListVodDomainsInput = vod20260101.ListVodDomainsInput

type ListVodDomainsOutput = vod20260101.ListVodDomainsOutput

type ListCdnDomainsParamForListVodDomainsInput = vod20260101.ListCdnDomainsParamForListVodDomainsInput
