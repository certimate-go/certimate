package obs

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
)

type PutBucketCustomDomainRequest struct {
	XMLName          xml.Name `json:"-"                          xml:"CustomDomainConfiguration"`
	CustomDomain     string   `json:"-"                          xml:"-"`
	Name             string   `json:"Name,omitempty"             xml:"Name,omitempty"`
	CertificateId    string   `json:"CertificateId,omitempty"    xml:"CertificateId,omitempty"`
	Certificate      string   `json:"Certificate,omitempty"      xml:"Certificate,omitempty"`
	CertificateChain string   `json:"CertificateChain,omitempty" xml:"CertificateChain,omitempty"`
	PrivateKey       string   `json:"PrivateKey,omitempty"       xml:"PrivateKey,omitempty"`
}

type PutBucketCustomDomainResponse struct {
	sdkResponseBase
}

func (c *Client) PutBucketCustomDomain(bucket string, req *PutBucketCustomDomainRequest) (*PutBucketCustomDomainResponse, error) {
	return c.PutBucketCustomDomainWithContext(context.Background(), bucket, req)
}

func (c *Client) PutBucketCustomDomainWithContext(ctx context.Context, bucket string, req *PutBucketCustomDomainRequest) (*PutBucketCustomDomainResponse, error) {
	if bucket == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset bucket")
	}
	if req.CustomDomain == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset customdomain")
	}

	httpreq, err := c.newRequest(http.MethodPut, "/?customdomain="+url.QueryEscape(req.CustomDomain), req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &PutBucketCustomDomainResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
