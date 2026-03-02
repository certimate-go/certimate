package internal

import (
	"context"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	aliesa "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/dara"
)

// This is a partial copy of https://github.com/alibabacloud-go/esa-20240910/blob/master/client/client_context_func.go
// to lightweight the vendor packages in the built binary.
type EsaClient struct {
	openapi.Client
	DisableSDKError *bool
}

func NewEsaClient(config *openapiutil.Config) (*EsaClient, error) {
	client := new(EsaClient)
	err := client.Init(config)
	return client, err
}

func (client *EsaClient) Init(config *openapiutil.Config) (_err error) {
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

func (client *EsaClient) ListCustomHostnamesWithContext(ctx context.Context, request *aliesa.ListCustomHostnamesRequest, runtime *dara.RuntimeOptions) (_result *aliesa.ListCustomHostnamesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.Hostname) {
		query["Hostname"] = request.Hostname
	}

	if !dara.IsNil(request.NameMatchType) {
		query["NameMatchType"] = request.NameMatchType
	}

	if !dara.IsNil(request.PageNumber) {
		query["PageNumber"] = request.PageNumber
	}

	if !dara.IsNil(request.PageSize) {
		query["PageSize"] = request.PageSize
	}

	if !dara.IsNil(request.RecordId) {
		query["RecordId"] = request.RecordId
	}

	if !dara.IsNil(request.SiteId) {
		query["SiteId"] = request.SiteId
	}

	if !dara.IsNil(request.Status) {
		query["Status"] = request.Status
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListCustomHostnames"),
		Version:     dara.String("2024-09-10"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliesa.ListCustomHostnamesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *EsaClient) UpdateCustomHostnameWithContext(ctx context.Context, request *aliesa.UpdateCustomHostnameRequest, runtime *dara.RuntimeOptions) (_result *aliesa.UpdateCustomHostnameResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.CasId) {
		query["CasId"] = request.CasId
	}

	if !dara.IsNil(request.CasRegion) {
		query["CasRegion"] = request.CasRegion
	}

	if !dara.IsNil(request.CertType) {
		query["CertType"] = request.CertType
	}

	if !dara.IsNil(request.Certificate) {
		query["Certificate"] = request.Certificate
	}

	if !dara.IsNil(request.HostnameId) {
		query["HostnameId"] = request.HostnameId
	}

	if !dara.IsNil(request.PrivateKey) {
		query["PrivateKey"] = request.PrivateKey
	}

	if !dara.IsNil(request.RecordId) {
		query["RecordId"] = request.RecordId
	}

	if !dara.IsNil(request.SslFlag) {
		query["SslFlag"] = request.SslFlag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateCustomHostname"),
		Version:     dara.String("2024-09-10"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliesa.UpdateCustomHostnameResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
