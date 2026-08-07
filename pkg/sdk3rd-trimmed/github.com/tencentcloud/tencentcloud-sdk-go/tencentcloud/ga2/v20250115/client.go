package v20250115

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ga2 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ga2/v20250115"
)

const APIVersion = ga2.APIVersion

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
	return ga2.NewDescribeListenersRequest()
}

func NewDescribeListenersResponse() (response *DescribeListenersResponse) {
	return ga2.NewDescribeListenersResponse()
}

func (c *Client) DescribeListenersWithContext(ctx context.Context, request *DescribeListenersRequest) (response *DescribeListenersResponse, err error) {
	if request == nil {
		request = NewDescribeListenersRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ga2", APIVersion, "DescribeListeners")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeListeners require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeListenersResponse()
	err = c.Send(request, response)
	return
}

func NewModifyListenerRequest() (request *ModifyListenerRequest) {
	return ga2.NewModifyListenerRequest()
}

func NewModifyListenerResponse() (response *ModifyListenerResponse) {
	return ga2.NewModifyListenerResponse()
}

func (c *Client) ModifyListenerWithContext(ctx context.Context, request *ModifyListenerRequest) (response *ModifyListenerResponse, err error) {
	if request == nil {
		request = NewModifyListenerRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "ga2", APIVersion, "ModifyListener")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyListener require credential")
	}

	request.SetContext(ctx)

	response = NewModifyListenerResponse()
	err = c.Send(request, response)
	return
}
