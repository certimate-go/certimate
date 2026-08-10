package huaweiibmc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ImportCustomCertificateToManagerRequest struct {
	sdkResponseBase

	ManagerID       string `json:"-"`
	ManagerLocation string `json:"-"`
	Certificate     string `json:"Certificate,omitempty"`
	Password        string `json:"Password,omitempty"`
}

type ImportCustomCertificateToManagerResponse struct {
	sdkResponseBase
}

func (c *Client) ImportCustomCertificateToManager(req *ImportCustomCertificateToManagerRequest) (*ImportCustomCertificateToManagerResponse, error) {
	return c.ImportCustomCertificateToManagerWithContext(context.Background(), req)
}

func (c *Client) ImportCustomCertificateToManagerWithContext(ctx context.Context, req *ImportCustomCertificateToManagerRequest) (*ImportCustomCertificateToManagerResponse, error) {
	managerLoc := strings.TrimRight(req.ManagerLocation, "/")
	if managerLoc == "" && req.ManagerID != "" {
		managerLoc = "/redfish/v1/Managers/" + url.PathEscape(req.ManagerID)
	}
	if managerLoc == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset managerId")
	}

	httpreq, err := c.newRequest(http.MethodPost, managerLoc+"/SecurityService/HttpsCert/Actions/HttpsCert.ImportCustomCertificateToManager")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &ImportCustomCertificateToManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
