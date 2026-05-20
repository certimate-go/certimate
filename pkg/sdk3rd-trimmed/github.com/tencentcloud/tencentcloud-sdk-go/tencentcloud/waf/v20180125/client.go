package v20180125

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcwaf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"
)

const APIVersion = tcwaf.APIVersion

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

func NewDescribeDomainDetailsSaasRequest() (request *DescribeDomainDetailsSaasRequest) {
	return tcwaf.NewDescribeDomainDetailsSaasRequest()
}

func NewDescribeDomainDetailsSaasResponse() (response *DescribeDomainDetailsSaasResponse) {
	return tcwaf.NewDescribeDomainDetailsSaasResponse()
}

func (c *Client) DescribeDomainDetailsSaasWithContext(ctx context.Context, request *DescribeDomainDetailsSaasRequest) (response *DescribeDomainDetailsSaasResponse, err error) {
	if request == nil {
		request = NewDescribeDomainDetailsSaasRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "waf", APIVersion, "DescribeDomainDetailsSaas")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeDomainDetailsSaas require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeDomainDetailsSaasResponse()
	err = c.Send(request, response)
	return
}

func (c *Client) ModifySpartaProtection(request *tcwaf.ModifySpartaProtectionRequest) (response *tcwaf.ModifySpartaProtectionResponse, err error) {
	return c.ModifySpartaProtectionWithContext(context.Background(), request)
}

func NewModifySpartaProtectionRequest() (request *ModifySpartaProtectionRequest) {
	return tcwaf.NewModifySpartaProtectionRequest()
}

func NewModifySpartaProtectionResponse() (response *ModifySpartaProtectionResponse) {
	return tcwaf.NewModifySpartaProtectionResponse()
}

func (c *Client) ModifySpartaProtectionWithContext(ctx context.Context, request *ModifySpartaProtectionRequest) (response *ModifySpartaProtectionResponse, err error) {
	if request == nil {
		request = NewModifySpartaProtectionRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "waf", APIVersion, "ModifySpartaProtection")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifySpartaProtection require credential")
	}

	request.SetContext(ctx)

	response = NewModifySpartaProtectionResponse()
	err = c.Send(request, response)
	return
}
