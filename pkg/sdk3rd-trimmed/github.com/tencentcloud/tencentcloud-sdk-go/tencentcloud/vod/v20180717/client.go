package v20180717

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod/v20180717"
)

const APIVersion = vod.APIVersion

type Client struct {
	common.Client
}

func NewDescribeVodDomainsRequest() (request *DescribeVodDomainsRequest) {
	return vod.NewDescribeVodDomainsRequest()
}

func NewDescribeVodDomainsResponse() (response *DescribeVodDomainsResponse) {
	return vod.NewDescribeVodDomainsResponse()
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func (c *Client) DescribeVodDomainsWithContext(ctx context.Context, request *DescribeVodDomainsRequest) (response *DescribeVodDomainsResponse, err error) {
	if request == nil {
		request = NewDescribeVodDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "vod", APIVersion, "DescribeVodDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeVodDomains require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeVodDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewSetVodDomainCertificateRequest() (request *SetVodDomainCertificateRequest) {
	return vod.NewSetVodDomainCertificateRequest()
}

func NewSetVodDomainCertificateResponse() (response *SetVodDomainCertificateResponse) {
	return vod.NewSetVodDomainCertificateResponse()
}

func (c *Client) SetVodDomainCertificateWithContext(ctx context.Context, request *SetVodDomainCertificateRequest) (response *SetVodDomainCertificateResponse, err error) {
	if request == nil {
		request = NewSetVodDomainCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "vod", APIVersion, "SetVodDomainCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("SetVodDomainCertificate require credential")
	}

	request.SetContext(ctx)

	response = NewSetVodDomainCertificateResponse()
	err = c.Send(request, response)
	return
}
