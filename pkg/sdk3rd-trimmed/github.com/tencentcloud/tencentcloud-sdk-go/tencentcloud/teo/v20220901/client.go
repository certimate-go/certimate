package v20220901

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

const APIVersion = teo.APIVersion

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

func NewDescribeAccelerationDomainsRequest() (request *DescribeAccelerationDomainsRequest) {
	return teo.NewDescribeAccelerationDomainsRequest()
}

func NewDescribeAccelerationDomainsResponse() (response *DescribeAccelerationDomainsResponse) {
	return teo.NewDescribeAccelerationDomainsResponse()
}

func (c *Client) DescribeAccelerationDomainsWithContext(ctx context.Context, request *DescribeAccelerationDomainsRequest) (response *DescribeAccelerationDomainsResponse, err error) {
	if request == nil {
		request = NewDescribeAccelerationDomainsRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "teo", APIVersion, "DescribeAccelerationDomains")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeAccelerationDomains require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeAccelerationDomainsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostCertificatesRequest() (request *DescribeHostCertificatesRequest) {
	request = &DescribeHostCertificatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("teo", APIVersion, "DescribeHostCertificates")

	return
}

func NewDescribeHostCertificatesResponse() (response *DescribeHostCertificatesResponse) {
	response = &DescribeHostCertificatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}

	return
}

func (c *Client) DescribeHostCertificatesWithContext(ctx context.Context, request *DescribeHostCertificatesRequest) (response *DescribeHostCertificatesResponse, err error) {
	if request == nil {
		request = NewDescribeHostCertificatesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "teo", APIVersion, "DescribeHostCertificates")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHostCertificates require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeHostCertificatesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyHostsCertificateRequest() (request *ModifyHostsCertificateRequest) {
	return teo.NewModifyHostsCertificateRequest()
}

func NewModifyHostsCertificateResponse() (response *ModifyHostsCertificateResponse) {
	return teo.NewModifyHostsCertificateResponse()
}

func (c *Client) ModifyHostsCertificateWithContext(ctx context.Context, request *ModifyHostsCertificateRequest) (response *ModifyHostsCertificateResponse, err error) {
	if request == nil {
		request = NewModifyHostsCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "teo", APIVersion, "ModifyHostsCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyHostsCertificate require credential")
	}

	request.SetContext(ctx)

	response = NewModifyHostsCertificateResponse()
	err = c.Send(request, response)
	return
}
