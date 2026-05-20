package cert

import (
	"fmt"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/http"
	"github.com/baidubce/bce-sdk-go/services/cert"
)

func (c *Client) CreateCert(args *CreateCertArgs) (*CreateCertResult, error) {
	if args == nil {
		return nil, fmt.Errorf("unset args")
	}

	result, err := c.Client.CreateCert(&args.CreateCertArgs)
	if err != nil {
		return nil, err
	}

	return &CreateCertResult{CreateCertResult: *result}, nil
}

func (c *Client) ListCerts() (*ListCertResult, error) {
	result, err := c.Client.ListCerts()
	if err != nil {
		return nil, err
	}

	return &ListCertResult{ListCertResult: *result}, nil
}

func (c *Client) ListCertDetail() (*ListCertDetailResult, error) {
	result, err := c.Client.ListCertDetail()
	if err != nil {
		return nil, err
	}

	return &ListCertDetailResult{ListCertDetailResult: *result}, nil
}

func (c *Client) GetCertMeta(id string) (*CertificateMeta, error) {
	result, err := c.Client.GetCertMeta(id)
	if err != nil {
		return nil, err
	}

	return &CertificateMeta{CertificateMeta: *result}, nil
}

func (c *Client) GetCertDetail(id string) (*CertificateDetailMeta, error) {
	result, err := c.Client.GetCertDetail(id)
	if err != nil {
		return nil, err
	}

	return &CertificateDetailMeta{CertificateDetailMeta: *result}, nil
}

func (c *Client) GetCertRawData(id string) (*CertificateRawData, error) {
	result := &CertificateRawData{}
	err := bce.NewRequestBuilder(c).
		WithMethod(http.GET).
		WithURL(cert.URI_PREFIX + cert.REQUEST_CERT_URL + "/" + id + "/rawData").
		WithResult(result).
		Do()

	return result, err
}

func (c *Client) UpdateCertName(id string, args *UpdateCertNameArgs) error {
	if args == nil {
		return fmt.Errorf("unset args")
	}

	err := c.Client.UpdateCertName(id, &args.UpdateCertNameArgs)
	return err
}

func (c *Client) UpdateCertData(id string, args *UpdateCertDataArgs) error {
	if args == nil {
		return fmt.Errorf("unset args")
	}

	err := c.Client.UpdateCertData(id, &args.UpdateCertDataArgs)
	return err
}

func (c *Client) DeleteCert(id string) error {
	err := c.Client.DeleteCert(id)
	return err
}
