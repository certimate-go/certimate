package v20180529

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcgaap "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"
)

const APIVersion = tcgaap.APIVersion

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

func NewDescribeHTTPSListenersRequest() (request *DescribeHTTPSListenersRequest) {
	return tcgaap.NewDescribeHTTPSListenersRequest()
}

func NewDescribeHTTPSListenersResponse() (response *DescribeHTTPSListenersResponse) {
	return tcgaap.NewDescribeHTTPSListenersResponse()
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
	return tcgaap.NewModifyHTTPSListenerAttributeRequest()
}

func NewModifyHTTPSListenerAttributeResponse() (response *ModifyHTTPSListenerAttributeResponse) {
	return tcgaap.NewModifyHTTPSListenerAttributeResponse()
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
