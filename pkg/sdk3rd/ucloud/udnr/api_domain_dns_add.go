package udnr

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DomainDNSAddRequest struct {
	request.CommonBase

	Dn         *string `required:"true"`
	RecordName *string `required:"true"`
	DnsType    *string `required:"true"`
	Content    *string `required:"true"`
	TTL        *string `required:"false"`
	Prio       *string `required:"false"`
}

type DomainDNSAddResponse struct {
	response.CommonBase
}

func (c *UDNRClient) NewDomainDNSAddRequest() *DomainDNSAddRequest {
	req := &DomainDNSAddRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UDNRClient) DomainDNSAdd(req *DomainDNSAddRequest) (*DomainDNSAddResponse, error) {
	var err error
	var res DomainDNSAddResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UdnrDomainDNSAdd", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
