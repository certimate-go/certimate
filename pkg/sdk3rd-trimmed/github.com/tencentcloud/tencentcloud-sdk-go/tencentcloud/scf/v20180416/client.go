package v20180416

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	scf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"
)

const APIVersion = scf.APIVersion

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

func NewGetCustomDomainRequest() (request *GetCustomDomainRequest) {
	return scf.NewGetCustomDomainRequest()
}

func NewGetCustomDomainResponse() (response *GetCustomDomainResponse) {
	return scf.NewGetCustomDomainResponse()
}

func (c *Client) GetCustomDomainWithContext(ctx context.Context, request *GetCustomDomainRequest) (response *GetCustomDomainResponse, err error) {
	if request == nil {
		request = NewGetCustomDomainRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "scf", APIVersion, "GetCustomDomain")

	if c.GetCredential() == nil {
		return nil, errors.New("GetCustomDomain require credential")
	}

	request.SetContext(ctx)

	response = NewGetCustomDomainResponse()
	err = c.Send(request, response)
	return
}

func NewListCustomDomainsRequest() (request *ListCustomDomainsRequest) {
	return scf.NewListCustomDomainsRequest()
}

func NewListCustomDomainsResponse() (response *ListCustomDomainsResponse) {
	return scf.NewListCustomDomainsResponse()
}

func (c *Client) ListCustomDomainsWithContext(ctx context.Context, request *ListCustomDomainsRequest) (response *ListCustomDomainsResponse, err error) {
	if request == nil {
		request = NewListCustomDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "scf", APIVersion, "ListCustomDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("ListCustomDomains require credential")
	}

	request.SetContext(ctx)

	response = NewListCustomDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateCustomDomainRequest() (request *UpdateCustomDomainRequest) {
	return scf.NewUpdateCustomDomainRequest()
}

func NewUpdateCustomDomainResponse() (response *UpdateCustomDomainResponse) {
	return scf.NewUpdateCustomDomainResponse()
}

func (c *Client) UpdateCustomDomainWithContext(ctx context.Context, request *UpdateCustomDomainRequest) (response *UpdateCustomDomainResponse, err error) {
	if request == nil {
		request = NewUpdateCustomDomainRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "scf", APIVersion, "UpdateCustomDomain")

	if c.GetCredential() == nil {
		return nil, errors.New("UpdateCustomDomain require credential")
	}

	request.SetContext(ctx)

	response = NewUpdateCustomDomainResponse()
	err = c.Send(request, response)
	return
}
