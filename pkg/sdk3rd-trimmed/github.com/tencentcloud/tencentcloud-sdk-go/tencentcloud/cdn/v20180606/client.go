package v20180606

import (
	"context"
	"errors"

	tccdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = tccdn.APIVersion

type Client struct {
	common.Client
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewDescribeCertDomainsRequest() (request *DescribeCertDomainsRequest) {
	return tccdn.NewDescribeCertDomainsRequest()
}

func NewDescribeCertDomainsResponse() (response *DescribeCertDomainsResponse) {
	return tccdn.NewDescribeCertDomainsResponse()
}

func (c *Client) DescribeCertDomainsWithContext(ctx context.Context, request *DescribeCertDomainsRequest) (response *DescribeCertDomainsResponse, err error) {
	if request == nil {
		request = NewDescribeCertDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "cdn", APIVersion, "DescribeCertDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeCertDomains require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeCertDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDomainsRequest() (request *DescribeDomainsRequest) {
	return tccdn.NewDescribeDomainsRequest()
}

func NewDescribeDomainsResponse() (response *DescribeDomainsResponse) {
	return tccdn.NewDescribeDomainsResponse()
}

func (c *Client) DescribeDomainsWithContext(ctx context.Context, request *DescribeDomainsRequest) (response *DescribeDomainsResponse, err error) {
	if request == nil {
		request = NewDescribeDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "cdn", APIVersion, "DescribeDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeDomains require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDomainsConfigRequest() (request *DescribeDomainsConfigRequest) {
	return tccdn.NewDescribeDomainsConfigRequest()
}

func NewDescribeDomainsConfigResponse() (response *DescribeDomainsConfigResponse) {
	return tccdn.NewDescribeDomainsConfigResponse()
}

func (c *Client) DescribeDomainsConfigWithContext(ctx context.Context, request *DescribeDomainsConfigRequest) (response *DescribeDomainsConfigResponse, err error) {
	if request == nil {
		request = NewDescribeDomainsConfigRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "cdn", APIVersion, "DescribeDomainsConfig")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeDomainsConfig require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeDomainsConfigResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateDomainConfigRequest() (request *UpdateDomainConfigRequest) {
	return tccdn.NewUpdateDomainConfigRequest()
}

func NewUpdateDomainConfigResponse() (response *UpdateDomainConfigResponse) {
	return tccdn.NewUpdateDomainConfigResponse()
}

func (c *Client) UpdateDomainConfigWithContext(ctx context.Context, request *UpdateDomainConfigRequest) (response *UpdateDomainConfigResponse, err error) {
	if request == nil {
		request = NewUpdateDomainConfigRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "cdn", APIVersion, "UpdateDomainConfig")

	if c.GetCredential() == nil {
		return nil, errors.New("UpdateDomainConfig require credential")
	}

	request.SetContext(ctx)

	response = NewUpdateDomainConfigResponse()
	err = c.Send(request, response)
	return
}
