package ulb

import (
	ucloudlb "github.com/ucloud/ucloud-sdk-go/services/ulb"
)

type UpdateListenerAttributeRequest = ucloudlb.UpdateListenerAttributeRequest

type UpdateListenerAttributeResponse = ucloudlb.UpdateListenerAttributeResponse

func (c *ULBClient) NewUpdateListenerAttributeRequest() *UpdateListenerAttributeRequest {
	req := &UpdateListenerAttributeRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *ULBClient) UpdateListenerAttribute(req *UpdateListenerAttributeRequest) (*UpdateListenerAttributeResponse, error) {
	var err error
	var res UpdateListenerAttributeResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UpdateListenerAttribute", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
