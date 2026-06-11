package v20180529

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	gaap "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"
)

const APIVersion = gaap.APIVersion

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

func NewCreateCertificateRequest() (request *CreateCertificateRequest) {
	return gaap.NewCreateCertificateRequest()
}

func NewCreateCertificateResponse() (response *CreateCertificateResponse) {
	return gaap.NewCreateCertificateResponse()
}

func (c *Client) CreateCertificateWithContext(ctx context.Context, request *CreateCertificateRequest) (response *CreateCertificateResponse, err error) {
	if request == nil {
		request = NewCreateCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "gaap", APIVersion, "CreateCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("CreateCertificate require credential")
	}

	request.SetContext(ctx)

	response = NewCreateCertificateResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCertificatesRequest() (request *DescribeCertificatesRequest) {
	return gaap.NewDescribeCertificatesRequest()
}

func NewDescribeCertificatesResponse() (response *DescribeCertificatesResponse) {
	return gaap.NewDescribeCertificatesResponse()
}

func (c *Client) DescribeCertificatesWithContext(ctx context.Context, request *DescribeCertificatesRequest) (response *DescribeCertificatesResponse, err error) {
	if request == nil {
		request = NewDescribeCertificatesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "gaap", APIVersion, "DescribeCertificates")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeCertificates require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeCertificatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCertificateDetailRequest() (request *DescribeCertificateDetailRequest) {
	return gaap.NewDescribeCertificateDetailRequest()
}

func NewDescribeCertificateDetailResponse() (response *DescribeCertificateDetailResponse) {
	return gaap.NewDescribeCertificateDetailResponse()
}

func (c *Client) DescribeCertificateDetailWithContext(ctx context.Context, request *DescribeCertificateDetailRequest) (response *DescribeCertificateDetailResponse, err error) {
	if request == nil {
		request = NewDescribeCertificateDetailRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "gaap", APIVersion, "DescribeCertificateDetail")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeCertificateDetail require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeCertificateDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHTTPSListenersRequest() (request *DescribeHTTPSListenersRequest) {
	return gaap.NewDescribeHTTPSListenersRequest()
}

func NewDescribeHTTPSListenersResponse() (response *DescribeHTTPSListenersResponse) {
	return gaap.NewDescribeHTTPSListenersResponse()
}

func (c *Client) DescribeHTTPSListenersWithContext(ctx context.Context, request *DescribeHTTPSListenersRequest) (response *DescribeHTTPSListenersResponse, err error) {
	if request == nil {
		request = NewDescribeHTTPSListenersRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "gaap", APIVersion, "DescribeHTTPSListeners")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHTTPSListeners require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeHTTPSListenersResponse()
	err = c.Send(request, response)
	return
}

func NewModifyHTTPSListenerAttributeRequest() (request *ModifyHTTPSListenerAttributeRequest) {
	return gaap.NewModifyHTTPSListenerAttributeRequest()
}

func NewModifyHTTPSListenerAttributeResponse() (response *ModifyHTTPSListenerAttributeResponse) {
	return gaap.NewModifyHTTPSListenerAttributeResponse()
}

func (c *Client) ModifyHTTPSListenerAttributeWithContext(ctx context.Context, request *ModifyHTTPSListenerAttributeRequest) (response *ModifyHTTPSListenerAttributeResponse, err error) {
	if request == nil {
		request = NewModifyHTTPSListenerAttributeRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "gaap", APIVersion, "ModifyHTTPSListenerAttribute")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyHTTPSListenerAttribute require credential")
	}

	request.SetContext(ctx)

	response = NewModifyHTTPSListenerAttributeResponse()
	err = c.Send(request, response)
	return
}
