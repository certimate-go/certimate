package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type DeleteSSLBindingRequest = ucloudlb.DeleteSSLBindingRequest

type DeleteSSLBindingResponse = ucloudlb.DeleteSSLBindingResponse

func (c *ULBClient) NewDeleteSSLBindingRequest() *DeleteSSLBindingRequest {
	req := &DeleteSSLBindingRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) DeleteSSLBinding(req *DeleteSSLBindingRequest) (*DeleteSSLBindingResponse, error) {
	var err error
	var res DeleteSSLBindingResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DeleteSSLBinding", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
