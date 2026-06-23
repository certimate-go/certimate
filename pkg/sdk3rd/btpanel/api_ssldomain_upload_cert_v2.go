package btpanel

import (
	"context"
	"net/http"
)

type SSLDomainUploadCertV2Request struct {
	PrivateKey  string `json:"key"`
	Certificate string `json:"cert"`
}

type SSLDomainUploadCertV2Response struct {
	sdkResponseBaseV2

	Message *struct {
		SSLHash string `json:"hash,omitempty"`
	} `json:"message,omitempty"`
}

func (c *Client) SSLDomainUploadCertV2(req *SSLDomainUploadCertV2Request) (*SSLDomainUploadCertV2Response, error) {
	return c.SSLDomainUploadCertV2WithContext(context.Background(), req)
}

func (c *Client) SSLDomainUploadCertV2WithContext(ctx context.Context, req *SSLDomainUploadCertV2Request) (*SSLDomainUploadCertV2Response, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/v2/ssl_domain?action=upload_cert", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SSLDomainUploadCertV2Response{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
