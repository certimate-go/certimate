package oss

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

type PutCnameRequest struct {
	XMLName xml.Name              `json:"-"          xml:"BucketCnameConfiguration"`
	Cname   *PutCnameRequestCname `json:",omitempty" xml:"Cname,omitempty"`
}

type PutCnameRequestCname struct {
	Domain                   *string                                       `json:",omitempty" xml:"Domain,omitempty"`
	CertificateConfiguration *PutCnameRequestCnameCertificateConfiguration `json:",omitempty" xml:"CertificateConfiguration,omitempty"`
}

type PutCnameRequestCnameCertificateConfiguration struct {
	CertId            *string `json:",omitempty" xml:"CertId,omitempty"`
	Certificate       *string `json:",omitempty" xml:"Certificate,omitempty"`
	PrivateKey        *string `json:",omitempty" xml:"PrivateKey,omitempty"`
	PreviousCertId    *string `json:",omitempty" xml:"PreviousCertId,omitempty"`
	Force             *bool   `json:",omitempty" xml:"Force,omitempty"`
	DeleteCertificate *bool   `json:",omitempty" xml:"DeleteCertificate,omitempty"`
}

type PutCnameResponse struct {
	sdkResponseBase
}

func (c *Client) PutBucketCname(req *PutCnameRequest) (*PutCnameResponse, error) {
	return c.PutBucketCnameWithContext(context.Background(), req)
}

func (c *Client) PutBucketCnameWithContext(ctx context.Context, req *PutCnameRequest) (*PutCnameResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/?cname&comp=add", req, fmt.Sprintf("/%s/", c.bucket))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &PutCnameResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
