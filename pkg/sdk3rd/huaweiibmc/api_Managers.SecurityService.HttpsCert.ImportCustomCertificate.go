package huaweiibmc

import (
	"context"
	"encoding/json"
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

func (r *ImportCustomCertificateToManagerResponse) GetAPIError() error {
	if r.isSuccessfulResponse() {
		return nil
	}

	return r.sdkResponseBase.GetAPIError()
}

func (r *ImportCustomCertificateToManagerResponse) isSuccessfulResponse() bool {
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
		if info.MessageID == "iBMC.1.0.CertImportOK" {
			return true
		}
	}

	return false
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

	httpreq, err := c.newRequest(http.MethodPost, managerLoc+"/SecurityService/HttpsCert/Actions/HttpsCert.ImportCustomCertificate")
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
