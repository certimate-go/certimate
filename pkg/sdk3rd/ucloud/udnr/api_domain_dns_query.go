package udnr

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DomainDNSQueryRequest struct {
	request.CommonBase

	Dn *string `required:"true"`
}

type DomainDNSQueryResponse struct {
	response.CommonBase

	Data []DomainDNSRecord
}

func (c *UDNRClient) NewDomainDNSQueryRequest() *DomainDNSQueryRequest {
	req := &DomainDNSQueryRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UDNRClient) DomainDNSQuery(req *DomainDNSQueryRequest) (*DomainDNSQueryResponse, error) {
	var err error
	var res DomainDNSQueryResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UdnrDomainDNSQuery", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
