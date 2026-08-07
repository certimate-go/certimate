package samwaf

import (
	"context"
	"net/http"
)

type VipConfigUploadSslCertRequest struct {
	CertContent string `json:"cert_content"`
	KeyContent  string `json:"key_content"`
}

type VipConfigUploadSslCertResponse struct {
	sdkResponseBase
}

func (c *Client) VipConfigUploadSslCert(req *VipConfigUploadSslCertRequest) (*VipConfigUploadSslCertResponse, error) {
	return c.VipConfigUploadSslCertWithContext(context.Background(), req)
}

func (c *Client) VipConfigUploadSslCertWithContext(ctx context.Context, req *VipConfigUploadSslCertRequest) (*VipConfigUploadSslCertResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/vipconfig/uploadSslCert")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &VipConfigUploadSslCertResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
