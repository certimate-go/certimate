package client

import (
	"context"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

func (client *Client) GetCustomDomainWithContext(ctx context.Context, domainName *string, headers map[string]*string, runtime *dara.RuntimeOptions) (_result *GetCustomDomainResponse, _err error) {
	req := &openapiutil.OpenApiRequest{
		Headers: headers,
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetCustomDomain"),
		Version:     dara.String("2023-03-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/2023-03-30/custom-domains/" + dara.PercentEncode(dara.StringValue(domainName))),
		Method:      dara.String("GET"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("ROA"),
		ReqBodyType: dara.String("json"),
		BodyType:    dara.String("json"),
	}
	_result = &GetCustomDomainResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListCustomDomainsWithContext(ctx context.Context, request *ListCustomDomainsRequest, headers map[string]*string, runtime *dara.RuntimeOptions) (_result *ListCustomDomainsResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.Limit) {
		query["limit"] = request.Limit
	}

	if !dara.IsNil(request.NextToken) {
		query["nextToken"] = request.NextToken
	}

	if !dara.IsNil(request.Prefix) {
		query["prefix"] = request.Prefix
	}

	req := &openapiutil.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListCustomDomains"),
		Version:     dara.String("2023-03-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/2023-03-30/custom-domains"),
		Method:      dara.String("GET"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("ROA"),
		ReqBodyType: dara.String("json"),
		BodyType:    dara.String("json"),
	}
	_result = &ListCustomDomainsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateCustomDomainWithContext(ctx context.Context, domainName *string, request *UpdateCustomDomainRequest, headers map[string]*string, runtime *dara.RuntimeOptions) (_result *UpdateCustomDomainResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	req := &openapiutil.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(request.Body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateCustomDomain"),
		Version:     dara.String("2023-03-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/2023-03-30/custom-domains/" + dara.PercentEncode(dara.StringValue(domainName))),
		Method:      dara.String("PUT"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("ROA"),
		ReqBodyType: dara.String("json"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateCustomDomainResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
