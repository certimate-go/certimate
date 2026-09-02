package v20200309

import (
	"context"
	"errors"

	antiddos "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/antiddos/v20200309"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = antiddos.APIVersion

type Client struct {
	common.Client
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewDescribeL7RulesBySSLCertIdRequest() (request *DescribeL7RulesBySSLCertIdRequest) {
	return antiddos.NewDescribeL7RulesBySSLCertIdRequest()
}

func NewDescribeL7RulesBySSLCertIdResponse() (response *DescribeL7RulesBySSLCertIdResponse) {
	return antiddos.NewDescribeL7RulesBySSLCertIdResponse()
}

func (c *Client) DescribeL7RulesBySSLCertIdWithContext(ctx context.Context, request *DescribeL7RulesBySSLCertIdRequest) (response *DescribeL7RulesBySSLCertIdResponse, err error) {
	if request == nil {
		request = NewDescribeL7RulesBySSLCertIdRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "antiddos", APIVersion, "DescribeL7RulesBySSLCertId")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeL7RulesBySSLCertId require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeL7RulesBySSLCertIdResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNewL7RulesRequest() (request *DescribeNewL7RulesRequest) {
	return antiddos.NewDescribeNewL7RulesRequest()
}

func NewDescribeNewL7RulesResponse() (response *DescribeNewL7RulesResponse) {
	return antiddos.NewDescribeNewL7RulesResponse()
}

func (c *Client) DescribeNewL7RulesWithContext(ctx context.Context, request *DescribeNewL7RulesRequest) (response *DescribeNewL7RulesResponse, err error) {
	if request == nil {
		request = NewDescribeNewL7RulesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "antiddos", APIVersion, "DescribeNewL7Rules")

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeNewL7Rules require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeNewL7RulesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNewDomainRulesRequest() (request *ModifyNewDomainRulesRequest) {
	return antiddos.NewModifyNewDomainRulesRequest()
}

func NewModifyNewDomainRulesResponse() (response *ModifyNewDomainRulesResponse) {
	return antiddos.NewModifyNewDomainRulesResponse()
}

func (c *Client) ModifyNewDomainRulesWithContext(ctx context.Context, request *ModifyNewDomainRulesRequest) (response *ModifyNewDomainRulesResponse, err error) {
	if request == nil {
		request = NewModifyNewDomainRulesRequest()
	}
	c.InitBaseRequest(&request.BaseRequest, "antiddos", APIVersion, "ModifyNewDomainRules")

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyNewDomainRules require credential")
	}

	request.SetContext(ctx)

	response = NewModifyNewDomainRulesResponse()
	err = c.Send(request, response)
	return
}
