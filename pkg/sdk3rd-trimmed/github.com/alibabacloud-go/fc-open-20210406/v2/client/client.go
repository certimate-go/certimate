package client

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Client struct {
	openapi.Client
}

func NewClient(config *openapi.Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *openapi.Config) (_err error) {
	_err = client.Client.Init(config)
	if _err != nil {
		return _err
	}
	_err = client.CheckConfig(config)
	if _err != nil {
		return _err
	}

	return nil
}

func (client *Client) GetCustomDomainWithOptions(domainName *string, headers *GetCustomDomainHeaders, runtime *util.RuntimeOptions) (_result *GetCustomDomainResponse, _err error) {
	realHeaders := make(map[string]*string)
	if !tea.BoolValue(util.IsUnset(headers.CommonHeaders)) {
		realHeaders = headers.CommonHeaders
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcAccountId)) {
		realHeaders["X-Fc-Account-Id"] = util.ToJSONString(headers.XFcAccountId)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcDate)) {
		realHeaders["X-Fc-Date"] = util.ToJSONString(headers.XFcDate)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcTraceId)) {
		realHeaders["X-Fc-Trace-Id"] = util.ToJSONString(headers.XFcTraceId)
	}

	req := &openapi.OpenApiRequest{
		Headers: realHeaders,
	}
	params := &openapi.Params{
		Action:      tea.String("GetCustomDomain"),
		Version:     tea.String("2021-04-06"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/2021-04-06/custom-domains/" + tea.StringValue(openapiutil.GetEncodeParam(domainName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetCustomDomainResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetCustomDomain(domainName *string) (_result *GetCustomDomainResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := &GetCustomDomainHeaders{}
	_result = &GetCustomDomainResponse{}
	_body, _err := client.GetCustomDomainWithOptions(domainName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListCustomDomainsWithOptions(request *ListCustomDomainsRequest, headers *ListCustomDomainsHeaders, runtime *util.RuntimeOptions) (_result *ListCustomDomainsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Limit)) {
		query["limit"] = request.Limit
	}

	if !tea.BoolValue(util.IsUnset(request.NextToken)) {
		query["nextToken"] = request.NextToken
	}

	if !tea.BoolValue(util.IsUnset(request.Prefix)) {
		query["prefix"] = request.Prefix
	}

	if !tea.BoolValue(util.IsUnset(request.StartKey)) {
		query["startKey"] = request.StartKey
	}

	realHeaders := make(map[string]*string)
	if !tea.BoolValue(util.IsUnset(headers.CommonHeaders)) {
		realHeaders = headers.CommonHeaders
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcAccountId)) {
		realHeaders["X-Fc-Account-Id"] = util.ToJSONString(headers.XFcAccountId)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcDate)) {
		realHeaders["X-Fc-Date"] = util.ToJSONString(headers.XFcDate)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcTraceId)) {
		realHeaders["X-Fc-Trace-Id"] = util.ToJSONString(headers.XFcTraceId)
	}

	req := &openapi.OpenApiRequest{
		Headers: realHeaders,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListCustomDomains"),
		Version:     tea.String("2021-04-06"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/2021-04-06/custom-domains"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListCustomDomainsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListCustomDomains(request *ListCustomDomainsRequest) (_result *ListCustomDomainsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := &ListCustomDomainsHeaders{}
	_result = &ListCustomDomainsResponse{}
	_body, _err := client.ListCustomDomainsWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateCustomDomainWithOptions(domainName *string, request *UpdateCustomDomainRequest, headers *UpdateCustomDomainHeaders, runtime *util.RuntimeOptions) (_result *UpdateCustomDomainResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.CertConfig)) {
		body["certConfig"] = request.CertConfig
	}

	if !tea.BoolValue(util.IsUnset(request.Protocol)) {
		body["protocol"] = request.Protocol
	}

	if !tea.BoolValue(util.IsUnset(request.RouteConfig)) {
		body["routeConfig"] = request.RouteConfig
	}

	if !tea.BoolValue(util.IsUnset(request.TlsConfig)) {
		body["tlsConfig"] = request.TlsConfig
	}

	if !tea.BoolValue(util.IsUnset(request.WafConfig)) {
		body["wafConfig"] = request.WafConfig
	}

	realHeaders := make(map[string]*string)
	if !tea.BoolValue(util.IsUnset(headers.CommonHeaders)) {
		realHeaders = headers.CommonHeaders
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcAccountId)) {
		realHeaders["X-Fc-Account-Id"] = util.ToJSONString(headers.XFcAccountId)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcDate)) {
		realHeaders["X-Fc-Date"] = util.ToJSONString(headers.XFcDate)
	}

	if !tea.BoolValue(util.IsUnset(headers.XFcTraceId)) {
		realHeaders["X-Fc-Trace-Id"] = util.ToJSONString(headers.XFcTraceId)
	}

	req := &openapi.OpenApiRequest{
		Headers: realHeaders,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateCustomDomain"),
		Version:     tea.String("2021-04-06"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/2021-04-06/custom-domains/" + tea.StringValue(openapiutil.GetEncodeParam(domainName))),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &UpdateCustomDomainResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateCustomDomain(domainName *string, request *UpdateCustomDomainRequest) (_result *UpdateCustomDomainResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := &UpdateCustomDomainHeaders{}
	_result = &UpdateCustomDomainResponse{}
	_body, _err := client.UpdateCustomDomainWithOptions(domainName, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}
