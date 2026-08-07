package elb

import (
	"context"
	"net/http"
)

type UpdateListenerRequest struct {
	ClientToken         *string `json:"clientToken,omitempty"`
	RegionID            *string `json:"regionID,omitempty"`
	ListenerID          *string `json:"listenerID,omitempty"`
	Name                *string `json:"name,omitempty"`
	Description         *string `json:"description,omitempty"`
	CertificateID       *string `json:"certificateID,omitempty"`
	CaEnabled           *bool   `json:"caEnabled,omitempty"`
	ClientCertificateID *string `json:"clientCertificateID,omitempty"`
}

type UpdateListenerResponse struct {
	sdkResponseBase

	ReturnObj []*Listener `json:"returnObj,omitempty"`
}

func (c *Client) UpdateListener(req *UpdateListenerRequest) (*UpdateListenerResponse, error) {
	return c.UpdateListenerWithContext(context.Background(), req)
}

func (c *Client) UpdateListenerWithContext(ctx context.Context, req *UpdateListenerRequest) (*UpdateListenerResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/v4/elb/update-listener")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateListenerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
