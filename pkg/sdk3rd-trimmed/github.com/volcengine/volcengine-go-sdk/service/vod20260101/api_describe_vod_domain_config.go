package vod20260101

import (
	"github.com/volcengine/volcengine-go-sdk/service/vod20260101"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opDescribeVodDomainConfig = "DescribeVodDomainConfig"

func (c *VOD20260101) DescribeVodDomainConfigRequest(input *DescribeVodDomainConfigInput) (req *request.Request, output *DescribeVodDomainConfigOutput) {
	op := &request.Operation{
		Name:       opDescribeVodDomainConfig,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeVodDomainConfigInput{}
	}

	output = &DescribeVodDomainConfigOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *VOD20260101) DescribeVodDomainConfigWithContext(ctx volcengine.Context, input *DescribeVodDomainConfigInput, opts ...request.Option) (*DescribeVodDomainConfigOutput, error) {
	req, out := c.DescribeVodDomainConfigRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type DescribeVodDomainConfigInput = vod20260101.DescribeVodDomainConfigInput

type DescribeVodDomainConfigOutput = vod20260101.DescribeVodDomainConfigOutput

type DescribeCdnDomainParamForDescribeVodDomainConfigInput = vod20260101.DescribeCdnDomainParamForDescribeVodDomainConfigInput
