package v20201207

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"
)

const APIVersion = tse.APIVersion

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

func NewCreateCloudNativeAPIGatewayCertificateRequest() (request *CreateCloudNativeAPIGatewayCertificateRequest) {
	return tse.NewCreateCloudNativeAPIGatewayCertificateRequest()
}

func NewCreateCloudNativeAPIGatewayCertificateResponse() (response *CreateCloudNativeAPIGatewayCertificateResponse) {
	return tse.NewCreateCloudNativeAPIGatewayCertificateResponse()
}

func (c *Client) CreateCloudNativeAPIGatewayCertificateWithContext(ctx context.Context, request *CreateCloudNativeAPIGatewayCertificateRequest) (response *CreateCloudNativeAPIGatewayCertificateResponse, err error) {
	if request == nil {
		request = NewCreateCloudNativeAPIGatewayCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "tse", APIVersion, "CreateCloudNativeAPIGatewayCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("CreateCloudNativeAPIGatewayCertificate require credential")
	}

	request.SetContext(ctx)

	response = NewCreateCloudNativeAPIGatewayCertificateResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCloudNativeAPIGatewayCertificatesRequest() (request *DescribeCloudNativeAPIGatewayCertificatesRequest) {
	return tse.NewDescribeCloudNativeAPIGatewayCertificatesRequest()
}

func NewDescribeCloudNativeAPIGatewayCertificatesResponse() (response *DescribeCloudNativeAPIGatewayCertificatesResponse) {
	return tse.NewDescribeCloudNativeAPIGatewayCertificatesResponse()
}

func (c *Client) DescribeCloudNativeAPIGatewayCertificatesWithContext(ctx context.Context, request *DescribeCloudNativeAPIGatewayCertificatesRequest) (response *DescribeCloudNativeAPIGatewayCertificatesResponse, err error) {
	if request == nil {
		request = NewDescribeCloudNativeAPIGatewayCertificatesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "tse", APIVersion, "DescribeCloudNativeAPIGatewayCertificates")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeCloudNativeAPIGatewayCertificates require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeCloudNativeAPIGatewayCertificatesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyCloudNativeAPIGatewayCertificateRequest() (request *ModifyCloudNativeAPIGatewayCertificateRequest) {
	return tse.NewModifyCloudNativeAPIGatewayCertificateRequest()
}

func NewModifyCloudNativeAPIGatewayCertificateResponse() (response *ModifyCloudNativeAPIGatewayCertificateResponse) {
	return tse.NewModifyCloudNativeAPIGatewayCertificateResponse()
}

func (c *Client) ModifyCloudNativeAPIGatewayCertificateWithContext(ctx context.Context, request *ModifyCloudNativeAPIGatewayCertificateRequest) (response *ModifyCloudNativeAPIGatewayCertificateResponse, err error) {
	if request == nil {
		request = NewModifyCloudNativeAPIGatewayCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "tse", APIVersion, "ModifyCloudNativeAPIGatewayCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyCloudNativeAPIGatewayCertificate require credential")
	}

	request.SetContext(ctx)

	response = NewModifyCloudNativeAPIGatewayCertificateResponse()
	err = c.Send(request, response)
	return
}
