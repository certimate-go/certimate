package cdn

import (
	"github.com/volcengine/volcengine-go-sdk/service/cdn"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opListCdnDomains = "ListCdnDomains"

func (c *CDN) ListCdnDomainsRequest(input *ListCdnDomainsInput) (req *request.Request, output *ListCdnDomainsOutput) {
	op := &request.Operation{
		Name:       opListCdnDomains,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListCdnDomainsInput{}
	}

	output = &ListCdnDomainsOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *CDN) ListCdnDomainsWithContext(ctx volcengine.Context, input *ListCdnDomainsInput, opts ...request.Option) (*ListCdnDomainsOutput, error) {
	req, out := c.ListCdnDomainsRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type ListCdnDomainsInput = cdn.ListCdnDomainsInput

type ListCdnDomainsOutput = cdn.ListCdnDomainsOutput
