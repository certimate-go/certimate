package v20180801

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tclive "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/live/v20180801"
)

const APIVersion = tclive.APIVersion

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

func NewDescribeLiveDomainsRequest() (request *DescribeLiveDomainsRequest) {
	return tclive.NewDescribeLiveDomainsRequest()
}

func NewDescribeLiveDomainsResponse() (response *DescribeLiveDomainsResponse) {
	return tclive.NewDescribeLiveDomainsResponse()
}

func (c *Client) DescribeLiveDomainsWithContext(ctx context.Context, request *DescribeLiveDomainsRequest) (response *DescribeLiveDomainsResponse, err error) {
	if request == nil {
		request = NewDescribeLiveDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "live", APIVersion, "DescribeLiveDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeLiveDomains require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeLiveDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLiveDomainCertBindingsRequest() (request *ModifyLiveDomainCertBindingsRequest) {
	return tclive.NewModifyLiveDomainCertBindingsRequest()
}

func NewModifyLiveDomainCertBindingsResponse() (response *ModifyLiveDomainCertBindingsResponse) {
	return tclive.NewModifyLiveDomainCertBindingsResponse()
}

func (c *Client) ModifyLiveDomainCertBindingsWithContext(ctx context.Context, request *ModifyLiveDomainCertBindingsRequest) (response *ModifyLiveDomainCertBindingsResponse, err error) {
	if request == nil {
		request = NewModifyLiveDomainCertBindingsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "live", APIVersion, "ModifyLiveDomainCertBindings")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyLiveDomainCertBindings require credential")
	}

	request.SetContext(ctx)

	response = NewModifyLiveDomainCertBindingsResponse()
	err = c.Send(request, response)
	return
}
