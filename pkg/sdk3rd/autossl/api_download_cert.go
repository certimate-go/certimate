package autossl

import (
	"context"
	"fmt"
	"net/http"
)

type DownloadCertRequest struct {
	// 证书 ID，即 `GetCertList` 返回的 `certId`。
	Id string `json:"id"`
}

type DownloadCertResponse struct {
	sdkResponseBase

	Keypair *CertificateKeypair `json:"-"`
}

func (r *DownloadCertResponse) GetKeypair() *CertificateKeypair {
	return r.Keypair
}

// 根据证书 ID 下载证书。
//
// REF: https://www.auto-ssl.cn/docs/OpenAPI.html#downloadcert
func (c *Client) DownloadCert(req *DownloadCertRequest) (*DownloadCertResponse, error) {
	return c.DownloadCertWithContext(context.Background(), req)
}

func (c *Client) DownloadCertWithContext(ctx context.Context, req *DownloadCertRequest) (*DownloadCertResponse, error) {
	if req == nil || req.Id == "" {
		return nil, fmt.Errorf("sdkerr: unset certId")
	}

	httpreq, err := c.newRequest(http.MethodPost, "/v1/api/cert/downloadCert")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DownloadCertResponse{}
	resp, err := c.doRequestWithResult(httpreq, result)
	if err != nil {
		return result, err
	}

	keypair, err := parseCertificateKeypair(resp.Body())
	if err != nil {
		return result, err
	} else {
		result.Keypair = keypair
	}

	return result, nil
}
