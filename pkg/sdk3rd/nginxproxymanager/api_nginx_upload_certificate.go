package nginxproxymanager

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type NginxUploadCertificateRequest struct {
	CertificateMeta
}

type NginxUploadCertificateResponse struct {
	CertificateMeta
}

func (c *Client) NginxUploadCertificate(certId int64, req *NginxUploadCertificateRequest) (*NginxUploadCertificateResponse, error) {
	return c.NginxUploadCertificateWithContext(context.Background(), certId, req)
}

func (c *Client) NginxUploadCertificateWithContext(ctx context.Context, certId int64, req *NginxUploadCertificateRequest) (*NginxUploadCertificateResponse, error) {
	if certId == 0 {
		return nil, fmt.Errorf("sdkerr: unset certId")
	}

	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/nginx/certificates/%d/upload", certId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetFileReader("certificate", "certificate.pem", strings.NewReader(req.Certificate))
		httpreq.SetFileReader("certificate_key", "privkey.pem", strings.NewReader(req.CertificateKey))
		httpreq.SetFileReader("intermediate_certificate", "cabundle.pem", strings.NewReader(req.IntermediateCertificate))
		httpreq.SetContext(ctx)
	}

	result := &NginxUploadCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
