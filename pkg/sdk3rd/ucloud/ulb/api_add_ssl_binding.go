package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type AddSSLBindingRequest = ucloudlb.AddSSLBindingRequest

type AddSSLBindingResponse = ucloudlb.AddSSLBindingResponse

func (c *ULBClient) NewAddSSLBindingRequest() *AddSSLBindingRequest {
	req := &AddSSLBindingRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) AddSSLBinding(req *AddSSLBindingRequest) (*AddSSLBindingResponse, error) {
	var err error
	var res AddSSLBindingResponse

	reqCopier := *req

	err = c.Client.InvokeAction("AddSSLBinding", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
