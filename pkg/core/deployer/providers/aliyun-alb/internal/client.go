package internal

import (
	"context"

	alialb "github.com/alibabacloud-go/alb-20200616/v2/client"
	alicas "github.com/alibabacloud-go/cas-20200407/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

// This is a partial copy of https://github.com/alibabacloud-go/cas-20200407/blob/master/client/client_context_func.go
// to lightweight the vendor packages in the built binary.
type CasClient struct {
	openapi.Client
	DisableSDKError *bool
}

func NewCasClient(config *openapiutil.Config) (*CasClient, error) {
	client := new(CasClient)
	err := client.Init(config)
	return client, err
}

func (client *CasClient) Init(config *openapiutil.Config) (_err error) {
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

func (client *CasClient) GetCertificateDetailWithContext(ctx context.Context, request *alicas.GetCertificateDetailRequest, runtime *dara.RuntimeOptions) (_result *alicas.GetCertificateDetailResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.CertificateId) {
		query["CertificateId"] = request.CertificateId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetCertificateDetail"),
		Version:     dara.String("2020-04-07"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alicas.GetCertificateDetailResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

// This is a partial copy of https://github.com/alibabacloud-go/alb-20200616/blob/master/client/client_context_func.go
// to lightweight the vendor packages in the built binary.
type AlbClient struct {
	openapi.Client
	DisableSDKError *bool
}

func NewAlbClient(config *openapiutil.Config) (*AlbClient, error) {
	client := new(AlbClient)
	err := client.Init(config)
	return client, err
}

func (client *AlbClient) Init(config *openapiutil.Config) (_err error) {
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

func (client *AlbClient) AssociateAdditionalCertificatesWithListenerWithContext(ctx context.Context, request *alialb.AssociateAdditionalCertificatesWithListenerRequest, runtime *dara.RuntimeOptions) (_result *alialb.AssociateAdditionalCertificatesWithListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.Certificates) {
		query["Certificates"] = request.Certificates
	}

	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("AssociateAdditionalCertificatesWithListener"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.AssociateAdditionalCertificatesWithListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) DissociateAdditionalCertificatesFromListenerWithContext(ctx context.Context, request *alialb.DissociateAdditionalCertificatesFromListenerRequest, runtime *dara.RuntimeOptions) (_result *alialb.DissociateAdditionalCertificatesFromListenerResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.Certificates) {
		query["Certificates"] = request.Certificates
	}

	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DissociateAdditionalCertificatesFromListener"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.DissociateAdditionalCertificatesFromListenerResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) GetListenerAttributeWithContext(ctx context.Context, request *alialb.GetListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alialb.GetListenerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetListenerAttribute"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.GetListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) GetLoadBalancerAttributeWithContext(ctx context.Context, request *alialb.GetLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alialb.GetLoadBalancerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetLoadBalancerAttribute"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.GetLoadBalancerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) ListListenerCertificatesWithContext(ctx context.Context, request *alialb.ListListenerCertificatesRequest, runtime *dara.RuntimeOptions) (_result *alialb.ListListenerCertificatesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.CertificateIds) {
		query["CertificateIds"] = request.CertificateIds
	}

	if !dara.IsNil(request.CertificateType) {
		query["CertificateType"] = request.CertificateType
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.MaxResults) {
		query["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		query["NextToken"] = request.NextToken
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListListenerCertificates"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.ListListenerCertificatesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) ListListenersWithContext(ctx context.Context, request *alialb.ListListenersRequest, runtime *dara.RuntimeOptions) (_result *alialb.ListListenersResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.ListenerIds) {
		query["ListenerIds"] = request.ListenerIds
	}

	if !dara.IsNil(request.ListenerProtocol) {
		query["ListenerProtocol"] = request.ListenerProtocol
	}

	if !dara.IsNil(request.LoadBalancerIds) {
		query["LoadBalancerIds"] = request.LoadBalancerIds
	}

	if !dara.IsNil(request.MaxResults) {
		query["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		query["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListListeners"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.ListListenersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *AlbClient) UpdateListenerAttributeWithContext(ctx context.Context, request *alialb.UpdateListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alialb.UpdateListenerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.CaCertificates) {
		query["CaCertificates"] = request.CaCertificates
	}

	if !dara.IsNil(request.CaEnabled) {
		query["CaEnabled"] = request.CaEnabled
	}

	if !dara.IsNil(request.Certificates) {
		query["Certificates"] = request.Certificates
	}

	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DefaultActions) {
		query["DefaultActions"] = request.DefaultActions
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.GzipEnabled) {
		query["GzipEnabled"] = request.GzipEnabled
	}

	if !dara.IsNil(request.Http2Enabled) {
		query["Http2Enabled"] = request.Http2Enabled
	}

	if !dara.IsNil(request.IdleTimeout) {
		query["IdleTimeout"] = request.IdleTimeout
	}

	if !dara.IsNil(request.ListenerDescription) {
		query["ListenerDescription"] = request.ListenerDescription
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.QuicConfig) {
		query["QuicConfig"] = request.QuicConfig
	}

	if !dara.IsNil(request.RequestTimeout) {
		query["RequestTimeout"] = request.RequestTimeout
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		query["SecurityPolicyId"] = request.SecurityPolicyId
	}

	if !dara.IsNil(request.XForwardedForConfig) {
		query["XForwardedForConfig"] = request.XForwardedForConfig
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateListenerAttribute"),
		Version:     dara.String("2020-06-16"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alialb.UpdateListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
