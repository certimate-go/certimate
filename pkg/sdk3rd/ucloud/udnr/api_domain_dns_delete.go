package udnr

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type DNSRecordDeleteRequest struct {
	request.CommonBase

	Dn         *string `required:"true"`
	RecordName *string `required:"true"`
	DnsType    *string `required:"true"`
	Content    *string `required:"true"`
}

type DNSRecordDeleteResponse struct {
	response.CommonBase
}

func (c *UDNRClient) NewDNSRecordDeleteRequest() *DNSRecordDeleteRequest {
	req := &DNSRecordDeleteRequest{}

	c.Client.SetupRequest(req)

	req.SetRetryable(true)
	return req
}

func (c *UDNRClient) DNSRecordDelete(req *DNSRecordDeleteRequest) (*DNSRecordDeleteResponse, error) {
	var err error
	var res DNSRecordDeleteResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UdnrDeleteDnsRecord", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
