package client

import (
	"context"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
)

func (client *Client) GetListenerAttributeWithContext(ctx context.Context, request *GetListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetListenerAttributeResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.ListenerId) {
		query["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetListenerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetLoadBalancerAttributeWithContext(ctx context.Context, request *GetLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *GetLoadBalancerAttributeResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.ClientToken) {
		query["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.DryRun) {
		query["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("GetLoadBalancerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &GetLoadBalancerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListListenersWithContext(ctx context.Context, request *ListListenersRequest, runtime *dara.RuntimeOptions) (_result *ListListenersResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = request.Validate()
		if _err != nil {
			return _result, _err
		}
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

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecSensorEnabled) {
		query["SecSensorEnabled"] = request.SecSensorEnabled
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ListListeners"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &ListListenersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateListenerAttributeWithContext(ctx context.Context, tmpReq *UpdateListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *UpdateListenerAttributeResponse, _err error) {
	if dara.BoolValue(client.EnableValidate) == true {
		_err = tmpReq.Validate()
		if _err != nil {
			return _result, _err
		}
	}
	request := &UpdateListenerAttributeShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !dara.IsNil(tmpReq.ProxyProtocolV2Config) {
		request.ProxyProtocolV2ConfigShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.ProxyProtocolV2Config, dara.String("ProxyProtocolV2Config"), dara.String("json"))
	}

	body := map[string]interface{}{}
	if !dara.IsNil(request.AlpnEnabled) {
		body["AlpnEnabled"] = request.AlpnEnabled
	}

	if !dara.IsNil(request.AlpnPolicy) {
		body["AlpnPolicy"] = request.AlpnPolicy
	}

	if !dara.IsNil(request.CaCertificateIds) {
		body["CaCertificateIds"] = request.CaCertificateIds
	}

	if !dara.IsNil(request.CaEnabled) {
		body["CaEnabled"] = request.CaEnabled
	}

	if !dara.IsNil(request.CertificateIds) {
		body["CertificateIds"] = request.CertificateIds
	}

	if !dara.IsNil(request.ClientToken) {
		body["ClientToken"] = request.ClientToken
	}

	if !dara.IsNil(request.Cps) {
		body["Cps"] = request.Cps
	}

	if !dara.IsNil(request.DryRun) {
		body["DryRun"] = request.DryRun
	}

	if !dara.IsNil(request.IdleTimeout) {
		body["IdleTimeout"] = request.IdleTimeout
	}

	if !dara.IsNil(request.ListenerDescription) {
		body["ListenerDescription"] = request.ListenerDescription
	}

	if !dara.IsNil(request.ListenerId) {
		body["ListenerId"] = request.ListenerId
	}

	if !dara.IsNil(request.Mss) {
		body["Mss"] = request.Mss
	}

	if !dara.IsNil(request.ProxyProtocolEnabled) {
		body["ProxyProtocolEnabled"] = request.ProxyProtocolEnabled
	}

	if !dara.IsNil(request.ProxyProtocolV2ConfigShrink) {
		body["ProxyProtocolV2Config"] = request.ProxyProtocolV2ConfigShrink
	}

	if !dara.IsNil(request.RegionId) {
		body["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.SecSensorEnabled) {
		body["SecSensorEnabled"] = request.SecSensorEnabled
	}

	if !dara.IsNil(request.SecurityPolicyId) {
		body["SecurityPolicyId"] = request.SecurityPolicyId
	}

	if !dara.IsNil(request.ServerGroupId) {
		body["ServerGroupId"] = request.ServerGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UpdateListenerAttribute"),
		Version:     dara.String("2022-04-30"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &UpdateListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
