package v20180317

import (
	"context"
	"errors"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = clb.APIVersion

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

func NewDescribeListenersRequest() (request *DescribeListenersRequest) {
	return clb.NewDescribeListenersRequest()
}

func NewDescribeListenersResponse() (response *DescribeListenersResponse) {
	return clb.NewDescribeListenersResponse()
}

func (c *Client) DescribeListenersWithContext(ctx context.Context, request *DescribeListenersRequest) (response *DescribeListenersResponse, err error) {
	if request == nil {
		request = NewDescribeListenersRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "clb", APIVersion, "DescribeListeners")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeListeners require credential")
	}

	request.SetContext(ctx)
	response = NewDescribeListenersResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeTaskStatusRequest() (request *DescribeTaskStatusRequest) {
	return clb.NewDescribeTaskStatusRequest()
}

func NewDescribeTaskStatusResponse() (response *DescribeTaskStatusResponse) {
	return clb.NewDescribeTaskStatusResponse()
}

func (c *Client) DescribeTaskStatusWithContext(ctx context.Context, request *DescribeTaskStatusRequest) (response *DescribeTaskStatusResponse, err error) {
	if request == nil {
		request = NewDescribeTaskStatusRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "clb", APIVersion, "DescribeTaskStatus")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeTaskStatus require credential")
	}

	request.SetContext(ctx)
	response = NewDescribeTaskStatusResponse()
	err = c.Send(request, response)
	return
}

func NewModifyDomainAttributesRequest() (request *ModifyDomainAttributesRequest) {
	return clb.NewModifyDomainAttributesRequest()
}

func NewModifyDomainAttributesResponse() (response *ModifyDomainAttributesResponse) {
	return clb.NewModifyDomainAttributesResponse()
}

func (c *Client) ModifyDomainAttributesWithContext(ctx context.Context, request *ModifyDomainAttributesRequest) (response *ModifyDomainAttributesResponse, err error) {
	if request == nil {
		request = NewModifyDomainAttributesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "clb", APIVersion, "ModifyDomainAttributes")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyDomainAttributes require credential")
	}

	request.SetContext(ctx)
	response = NewModifyDomainAttributesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyListenerRequest() (request *ModifyListenerRequest) {
	return clb.NewModifyListenerRequest()
}

func NewModifyListenerResponse() (response *ModifyListenerResponse) {
	return clb.NewModifyListenerResponse()
}

func (c *Client) ModifyListenerWithContext(ctx context.Context, request *ModifyListenerRequest) (response *ModifyListenerResponse, err error) {
	if request == nil {
		request = NewModifyListenerRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "clb", APIVersion, "ModifyListener")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyListener require credential")
	}

	request.SetContext(ctx)
	response = NewModifyListenerResponse()
	err = c.Send(request, response)
	return
}
