package huaweiibmc

import (
	"context"
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
