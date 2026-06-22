package vod20260101

import (
	"github.com/volcengine/volcengine-go-sdk/service/vod20260101"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/request"
)

const opUpdateVodDomainConfig = "UpdateVodDomainConfig"

func (c *VOD20260101) UpdateVodDomainConfigRequest(input *UpdateVodDomainConfigInput) (req *request.Request, output *UpdateVodDomainConfigOutput) {
	op := &request.Operation{
		Name:       opUpdateVodDomainConfig,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &UpdateVodDomainConfigInput{}
	}

	output = &UpdateVodDomainConfigOutput{}
	req = c.newRequest(op, input, output)

	req.HTTPRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	return
}

func (c *VOD20260101) UpdateVodDomainConfigWithContext(ctx volcengine.Context, input *UpdateVodDomainConfigInput, opts ...request.Option) (*UpdateVodDomainConfigOutput, error) {
	req, out := c.UpdateVodDomainConfigRequest(input)
	req.SetContext(ctx)
	req.ApplyOptions(opts...)
	return out, req.Send()
}

type UpdateVodDomainConfigInput = vod20260101.UpdateVodDomainConfigInput

type UpdateVodDomainConfigOutput = vod20260101.UpdateVodDomainConfigOutput

type UpdateCdnConfigParamForUpdateVodDomainConfigInput = vod20260101.UpdateCdnConfigParamForUpdateVodDomainConfigInput

type HTTPSForUpdateVodDomainConfigInput = vod20260101.HTTPSForUpdateVodDomainConfigInput

type CertInfoForUpdateVodDomainConfigInput = vod20260101.CertInfoForUpdateVodDomainConfigInput
