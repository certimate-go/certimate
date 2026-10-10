package client

import (
	"context"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

func (client *Client) ListGatewayDomainWithContext(ctx context.Context, request *ListGatewayDomainRequest, runtime *dara.RuntimeOptions) (_result *ListGatewayDomainResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.AcceptLanguage) {
		query["AcceptLanguage"] = request.AcceptLanguage
	}

	if !dara.IsNil(request.DomainName) {
		query["DomainName"] = request.DomainName
	}

	if !dara.IsNil(request.GatewayUniqueId) {
		query["GatewayUniqueId"] = request.GatewayUniqueId
	}

	if !dara.IsNil(request.Type) {
		query["Type"] = request.Type
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListGatewayDomain"),
		Version:     dara.String("2019-05-31"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListGatewayDomainResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateGatewayDomainWithContext(ctx context.Context, tmpReq *UpdateGatewayDomainRequest, runtime *dara.RuntimeOptions) (_result *UpdateGatewayDomainResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = tmpReq.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	request := &UpdateGatewayDomainShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !dara.IsNil(tmpReq.TlsCipherSuitesConfigJSON) {
		request.TlsCipherSuitesConfigJSONShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.TlsCipherSuitesConfigJSON, dara.String("TlsCipherSuitesConfigJSON"), dara.String("json"))
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.AcceptLanguage) {
		query["AcceptLanguage"] = request.AcceptLanguage
	}

	if !dara.IsNil(request.CaCertIdentifier) {
		query["CaCertIdentifier"] = request.CaCertIdentifier
	}

	if !dara.IsNil(request.CertIdentifier) {
		query["CertIdentifier"] = request.CertIdentifier
	}

	if !dara.IsNil(request.GatewayUniqueId) {
		query["GatewayUniqueId"] = request.GatewayUniqueId
	}

	if !dara.IsNil(request.Http2) {
		query["Http2"] = request.Http2
	}

	if !dara.IsNil(request.Id) {
		query["Id"] = request.Id
	}

	if !dara.IsNil(request.MtlsEnabled) {
		query["MtlsEnabled"] = request.MtlsEnabled
	}

	if !dara.IsNil(request.MustHttps) {
		query["MustHttps"] = request.MustHttps
	}

	if !dara.IsNil(request.Protocol) {
		query["Protocol"] = request.Protocol
	}

	if !dara.IsNil(request.TlsCipherSuitesConfigJSONShrink) {
		query["TlsCipherSuitesConfigJSON"] = request.TlsCipherSuitesConfigJSONShrink
	}

	if !dara.IsNil(request.TlsMax) {
		query["TlsMax"] = request.TlsMax
	}

	if !dara.IsNil(request.TlsMin) {
		query["TlsMin"] = request.TlsMin
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateGatewayDomain"),
		Version:     dara.String("2019-05-31"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateGatewayDomainResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateSSLCertWithContext(ctx context.Context, request *UpdateSSLCertRequest, runtime *dara.RuntimeOptions) (_result *UpdateSSLCertResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.AcceptLanguage) {
		query["AcceptLanguage"] = request.AcceptLanguage
	}

	if !dara.IsNil(request.CertIdentifier) {
		query["CertIdentifier"] = request.CertIdentifier
	}

	if !dara.IsNil(request.DomainId) {
		query["DomainId"] = request.DomainId
	}

	if !dara.IsNil(request.GatewayUniqueId) {
		query["GatewayUniqueId"] = request.GatewayUniqueId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateSSLCert"),
		Version:     dara.String("2019-05-31"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateSSLCertResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
