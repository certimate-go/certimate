package console

import (
	"context"
	"fmt"
	"net/http"
)

type GetHttpsServiceManagerResponse struct {
	sdkResponseBase
	Data *struct {
		sdkResponseBaseData
		Status  int32                       `json:"status"`
		Domains []HttpsServiceManagerDomain `json:"result"`
	} `json:"data,omitempty"`
}

type HttpsServiceManagerDomain struct {
	CertificateId string                            `json:"certificate_id"`
	CommonName    string                            `json:"commonName"`
	Https         bool                              `json:"https"`
	ForceHttps    bool                              `json:"force_https"`
	PaymentType   string                            `json:"payment_type"`
	DomainType    string                            `json:"domain_type"`
	Validity      HttpsServiceManagerDomainValidity `json:"validity"`
}

type HttpsServiceManagerDomainValidity struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

func (c *Client) GetHttpsServiceManager(domain string) (*GetHttpsServiceManagerResponse, error) {
	return c.GetHttpsServiceManagerWithContext(context.Background(), domain)
}

func (c *Client) GetHttpsServiceManagerWithContext(ctx context.Context, domain string) (*GetHttpsServiceManagerResponse, error) {
	if domain == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset domain")
	}

	if err := c.ensureCookies(ctx); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/api/https/services/manager")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetQueryParam("domain", domain)
		httpreq.SetContext(ctx)
	}

	result := &GetHttpsServiceManagerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
