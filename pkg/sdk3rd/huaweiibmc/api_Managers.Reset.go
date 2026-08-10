package huaweiibmc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ResetManagerRequest struct {
	sdkResponseBase

	ManagerID       string `json:"-"`
	ManagerLocation string `json:"-"`
	ResetType       string `json:"ResetType,omitempty"`
}

type ResetManagerResponse struct {
	sdkResponseBase
}

func (r *ResetManagerResponse) GetAPIError() error {
	if r.isSuccessfulResponse() {
		return nil
	}

	return r.sdkResponseBase.GetAPIError()
}

func (r *ResetManagerResponse) isSuccessfulResponse() bool {
	if r.Error == nil {
		return false
	}

	var extendedInfo []struct {
		MessageID string `json:"MessageId"`
	}
	if err := json.Unmarshal(r.Error.MessageExtendedInfo, &extendedInfo); err != nil {
		return false
	}

	for _, info := range extendedInfo {
		if info.MessageID == "Base.1.0.Success" {
			return true
		}
	}

	return false
}

func (c *Client) ResetManager(req *ResetManagerRequest) (*ResetManagerResponse, error) {
	return c.ResetManagerWithContext(context.Background(), req)
}

func (c *Client) ResetManagerWithContext(ctx context.Context, req *ResetManagerRequest) (*ResetManagerResponse, error) {
	managerLoc := strings.TrimRight(req.ManagerLocation, "/")
	if managerLoc == "" && req.ManagerID != "" {
		managerLoc = "/redfish/v1/Managers/" + url.PathEscape(req.ManagerID)
	}
	if managerLoc == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset managerId")
	}

	httpreq, err := c.newRequest(http.MethodPost, managerLoc+"/Actions/Manager.Reset")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ResetManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
