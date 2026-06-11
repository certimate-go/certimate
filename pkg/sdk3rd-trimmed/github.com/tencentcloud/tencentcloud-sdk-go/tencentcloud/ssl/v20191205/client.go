package v20191205

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

const APIVersion = ssl.APIVersion

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

func NewDescribeCertificateRequest() (request *DescribeCertificateRequest) {
	return ssl.NewDescribeCertificateRequest()
}

func NewDescribeCertificateResponse() (response *DescribeCertificateResponse) {
	return ssl.NewDescribeCertificateResponse()
}

func (c *Client) DescribeCertificateWithContext(ctx context.Context, request *DescribeCertificateRequest) (response *DescribeCertificateResponse, err error) {
	if request == nil {
		request = NewDescribeCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", ssl.APIVersion, "DescribeCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeCertificate require credential")
	}

	request.SetContext(ctx)
	response = NewDescribeCertificateResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostCosInstanceListRequest() (request *DescribeHostCosInstanceListRequest) {
	return ssl.NewDescribeHostCosInstanceListRequest()
}

func NewDescribeHostCosInstanceListResponse() (response *DescribeHostCosInstanceListResponse) {
	return ssl.NewDescribeHostCosInstanceListResponse()
}

func (c *Client) DescribeHostCosInstanceListWithContext(ctx context.Context, request *DescribeHostCosInstanceListRequest) (response *DescribeHostCosInstanceListResponse, err error) {
	if request == nil {
		request = NewDescribeHostCosInstanceListRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", ssl.APIVersion, "DescribeHostCosInstanceList")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHostCosInstanceList require credential")
	}

	request.SetContext(ctx)
	response = NewDescribeHostCosInstanceListResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostDeployRecordDetailRequest() (request *DescribeHostDeployRecordDetailRequest) {
	return ssl.NewDescribeHostDeployRecordDetailRequest()
}

func NewDescribeHostDeployRecordDetailResponse() (response *DescribeHostDeployRecordDetailResponse) {
	return ssl.NewDescribeHostDeployRecordDetailResponse()
}

func (c *Client) DescribeHostDeployRecordDetailWithContext(ctx context.Context, request *DescribeHostDeployRecordDetailRequest) (response *DescribeHostDeployRecordDetailResponse, err error) {
	if request == nil {
		request = NewDescribeHostDeployRecordDetailRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "DescribeHostDeployRecordDetail")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHostDeployRecordDetail require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeHostDeployRecordDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostUpdateRecordDetailRequest() (request *DescribeHostUpdateRecordDetailRequest) {
	return ssl.NewDescribeHostUpdateRecordDetailRequest()
}

func NewDescribeHostUpdateRecordDetailResponse() (response *DescribeHostUpdateRecordDetailResponse) {
	return ssl.NewDescribeHostUpdateRecordDetailResponse()
}

func (c *Client) DescribeHostUpdateRecordDetailWithContext(ctx context.Context, request *DescribeHostUpdateRecordDetailRequest) (response *DescribeHostUpdateRecordDetailResponse, err error) {
	if request == nil {
		request = NewDescribeHostUpdateRecordDetailRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "DescribeHostUpdateRecordDetail")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHostUpdateRecordDetail require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeHostUpdateRecordDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostUploadUpdateRecordDetailRequest() (request *DescribeHostUploadUpdateRecordDetailRequest) {
	return ssl.NewDescribeHostUploadUpdateRecordDetailRequest()
}

func NewDescribeHostUploadUpdateRecordDetailResponse() (response *DescribeHostUploadUpdateRecordDetailResponse) {
	return ssl.NewDescribeHostUploadUpdateRecordDetailResponse()
}

func (c *Client) DescribeHostUploadUpdateRecordDetailWithContext(ctx context.Context, request *DescribeHostUploadUpdateRecordDetailRequest) (response *DescribeHostUploadUpdateRecordDetailResponse, err error) {
	if request == nil {
		request = NewDescribeHostUploadUpdateRecordDetailRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "DescribeHostUploadUpdateRecordDetail")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeHostUploadUpdateRecordDetail require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeHostUploadUpdateRecordDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDeployCertificateInstanceRequest() (request *DeployCertificateInstanceRequest) {
	return ssl.NewDeployCertificateInstanceRequest()
}

func NewDeployCertificateInstanceResponse() (response *DeployCertificateInstanceResponse) {
	return ssl.NewDeployCertificateInstanceResponse()
}

func (c *Client) DeployCertificateInstanceWithContext(ctx context.Context, request *DeployCertificateInstanceRequest) (response *DeployCertificateInstanceResponse, err error) {
	if request == nil {
		request = NewDeployCertificateInstanceRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "DeployCertificateInstance")

	if c.GetCredential() == nil {
		return nil, errors.New("DeployCertificateInstance require credential")
	}

	request.SetContext(ctx)

	response = NewDeployCertificateInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateCertificateInstanceRequest() (request *UpdateCertificateInstanceRequest) {
	return ssl.NewUpdateCertificateInstanceRequest()
}

func NewUpdateCertificateInstanceResponse() (response *UpdateCertificateInstanceResponse) {
	return ssl.NewUpdateCertificateInstanceResponse()
}

func (c *Client) UpdateCertificateInstanceWithContext(ctx context.Context, request *UpdateCertificateInstanceRequest) (response *UpdateCertificateInstanceResponse, err error) {
	if request == nil {
		request = NewUpdateCertificateInstanceRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "UpdateCertificateInstance")

	if c.GetCredential() == nil {
		return nil, errors.New("UpdateCertificateInstance require credential")
	}

	request.SetContext(ctx)

	response = NewUpdateCertificateInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewUploadCertificateRequest() (request *UploadCertificateRequest) {
	return ssl.NewUploadCertificateRequest()
}

func NewUploadCertificateResponse() (response *UploadCertificateResponse) {
	return ssl.NewUploadCertificateResponse()
}

func (c *Client) UploadCertificateWithContext(ctx context.Context, request *UploadCertificateRequest) (response *UploadCertificateResponse, err error) {
	if request == nil {
		request = NewUploadCertificateRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "UploadCertificate")

	if c.GetCredential() == nil {
		return nil, errors.New("UploadCertificate require credential")
	}

	request.SetContext(ctx)
	response = NewUploadCertificateResponse()
	err = c.Send(request, response)
	return
}

func NewUploadUpdateCertificateInstanceRequest() (request *UploadUpdateCertificateInstanceRequest) {
	return ssl.NewUploadUpdateCertificateInstanceRequest()
}

func NewUploadUpdateCertificateInstanceResponse() (response *UploadUpdateCertificateInstanceResponse) {
	return ssl.NewUploadUpdateCertificateInstanceResponse()
}

func (c *Client) UploadUpdateCertificateInstanceWithContext(ctx context.Context, request *UploadUpdateCertificateInstanceRequest) (response *UploadUpdateCertificateInstanceResponse, err error) {
	if request == nil {
		request = NewUploadUpdateCertificateInstanceRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ssl", APIVersion, "UploadUpdateCertificateInstance")

	if c.GetCredential() == nil {
		return nil, errors.New("UploadUpdateCertificateInstance require credential")
	}

	request.SetContext(ctx)

	response = NewUploadUpdateCertificateInstanceResponse()
	err = c.Send(request, response)
	return
}
