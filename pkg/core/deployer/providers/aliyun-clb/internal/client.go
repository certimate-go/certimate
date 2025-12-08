package internal

import (
	"context"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	alislb "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/dara"
)

// This is a partial copy of https://github.com/alibabacloud-go/slb-20140515/blob/master/client/client_context_func.go
// to lightweight the vendor packages in the built binary.
type SlbClient struct {
	openapi.Client
}

func NewSlbClient(config *openapi.Config) (*SlbClient, error) {
	client := new(SlbClient)
	err := client.Init(config)
	return client, err
}

func (client *SlbClient) Init(config *openapi.Config) (_err error) {
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

func (client *SlbClient) DescribeDomainExtensionsWithContext(ctx context.Context, request *alislb.DescribeDomainExtensionsRequest, runtime *dara.RuntimeOptions) (_result *alislb.DescribeDomainExtensionsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.DomainExtensionId) {
		query["DomainExtensionId"] = request.DomainExtensionId
	}

	if !dara.IsNil(request.ListenerPort) {
		query["ListenerPort"] = request.ListenerPort
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeDomainExtensions"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.DescribeDomainExtensionsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) DescribeLoadBalancerListenersWithContext(ctx context.Context, request *alislb.DescribeLoadBalancerListenersRequest, runtime *dara.RuntimeOptions) (_result *alislb.DescribeLoadBalancerListenersResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.Description) {
		query["Description"] = request.Description
	}

	if !dara.IsNil(request.ListenerPort) {
		query["ListenerPort"] = request.ListenerPort
	}

	if !dara.IsNil(request.ListenerProtocol) {
		query["ListenerProtocol"] = request.ListenerProtocol
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.MaxResults) {
		query["MaxResults"] = request.MaxResults
	}

	if !dara.IsNil(request.NextToken) {
		query["NextToken"] = request.NextToken
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeLoadBalancerListeners"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.DescribeLoadBalancerListenersResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) DescribeLoadBalancerAttributeWithContext(ctx context.Context, request *alislb.DescribeLoadBalancerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alislb.DescribeLoadBalancerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeLoadBalancerAttribute"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.DescribeLoadBalancerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) DescribeLoadBalancerHTTPSListenerAttributeWithContext(ctx context.Context, request *alislb.DescribeLoadBalancerHTTPSListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alislb.DescribeLoadBalancerHTTPSListenerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.ListenerPort) {
		query["ListenerPort"] = request.ListenerPort
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeLoadBalancerHTTPSListenerAttribute"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.DescribeLoadBalancerHTTPSListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) SetDomainExtensionAttributeWithContext(ctx context.Context, request *alislb.SetDomainExtensionAttributeRequest, runtime *dara.RuntimeOptions) (_result *alislb.SetDomainExtensionAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.DomainExtensionId) {
		query["DomainExtensionId"] = request.DomainExtensionId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	if !dara.IsNil(request.ServerCertificateId) {
		query["ServerCertificateId"] = request.ServerCertificateId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("SetDomainExtensionAttribute"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.SetDomainExtensionAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) SetLoadBalancerHTTPSListenerAttributeWithContext(ctx context.Context, request *alislb.SetLoadBalancerHTTPSListenerAttributeRequest, runtime *dara.RuntimeOptions) (_result *alislb.SetLoadBalancerHTTPSListenerAttributeResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.AclId) {
		query["AclId"] = request.AclId
	}

	if !dara.IsNil(request.AclStatus) {
		query["AclStatus"] = request.AclStatus
	}

	if !dara.IsNil(request.AclType) {
		query["AclType"] = request.AclType
	}

	if !dara.IsNil(request.Bandwidth) {
		query["Bandwidth"] = request.Bandwidth
	}

	if !dara.IsNil(request.CACertificateId) {
		query["CACertificateId"] = request.CACertificateId
	}

	if !dara.IsNil(request.Cookie) {
		query["Cookie"] = request.Cookie
	}

	if !dara.IsNil(request.CookieTimeout) {
		query["CookieTimeout"] = request.CookieTimeout
	}

	if !dara.IsNil(request.Description) {
		query["Description"] = request.Description
	}

	if !dara.IsNil(request.EnableHttp2) {
		query["EnableHttp2"] = request.EnableHttp2
	}

	if !dara.IsNil(request.Gzip) {
		query["Gzip"] = request.Gzip
	}

	if !dara.IsNil(request.HealthCheck) {
		query["HealthCheck"] = request.HealthCheck
	}

	if !dara.IsNil(request.HealthCheckConnectPort) {
		query["HealthCheckConnectPort"] = request.HealthCheckConnectPort
	}

	if !dara.IsNil(request.HealthCheckDomain) {
		query["HealthCheckDomain"] = request.HealthCheckDomain
	}

	if !dara.IsNil(request.HealthCheckHttpCode) {
		query["HealthCheckHttpCode"] = request.HealthCheckHttpCode
	}

	if !dara.IsNil(request.HealthCheckInterval) {
		query["HealthCheckInterval"] = request.HealthCheckInterval
	}

	if !dara.IsNil(request.HealthCheckMethod) {
		query["HealthCheckMethod"] = request.HealthCheckMethod
	}

	if !dara.IsNil(request.HealthCheckTimeout) {
		query["HealthCheckTimeout"] = request.HealthCheckTimeout
	}

	if !dara.IsNil(request.HealthCheckURI) {
		query["HealthCheckURI"] = request.HealthCheckURI
	}

	if !dara.IsNil(request.HealthyThreshold) {
		query["HealthyThreshold"] = request.HealthyThreshold
	}

	if !dara.IsNil(request.IdleTimeout) {
		query["IdleTimeout"] = request.IdleTimeout
	}

	if !dara.IsNil(request.ListenerPort) {
		query["ListenerPort"] = request.ListenerPort
	}

	if !dara.IsNil(request.LoadBalancerId) {
		query["LoadBalancerId"] = request.LoadBalancerId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.RequestTimeout) {
		query["RequestTimeout"] = request.RequestTimeout
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	if !dara.IsNil(request.Scheduler) {
		query["Scheduler"] = request.Scheduler
	}

	if !dara.IsNil(request.ServerCertificateId) {
		query["ServerCertificateId"] = request.ServerCertificateId
	}

	if !dara.IsNil(request.StickySession) {
		query["StickySession"] = request.StickySession
	}

	if !dara.IsNil(request.StickySessionType) {
		query["StickySessionType"] = request.StickySessionType
	}

	if !dara.IsNil(request.TLSCipherPolicy) {
		query["TLSCipherPolicy"] = request.TLSCipherPolicy
	}

	if !dara.IsNil(request.UnhealthyThreshold) {
		query["UnhealthyThreshold"] = request.UnhealthyThreshold
	}

	if !dara.IsNil(request.VServerGroup) {
		query["VServerGroup"] = request.VServerGroup
	}

	if !dara.IsNil(request.VServerGroupId) {
		query["VServerGroupId"] = request.VServerGroupId
	}

	if !dara.IsNil(request.XForwardedFor) {
		query["XForwardedFor"] = request.XForwardedFor
	}

	if !dara.IsNil(request.XForwardedFor_ClientSrcPort) {
		query["XForwardedFor_ClientSrcPort"] = request.XForwardedFor_ClientSrcPort
	}

	if !dara.IsNil(request.XForwardedFor_SLBID) {
		query["XForwardedFor_SLBID"] = request.XForwardedFor_SLBID
	}

	if !dara.IsNil(request.XForwardedFor_SLBIP) {
		query["XForwardedFor_SLBIP"] = request.XForwardedFor_SLBIP
	}

	if !dara.IsNil(request.XForwardedFor_SLBPORT) {
		query["XForwardedFor_SLBPORT"] = request.XForwardedFor_SLBPORT
	}

	if !dara.IsNil(request.XForwardedFor_proto) {
		query["XForwardedFor_proto"] = request.XForwardedFor_proto
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("SetLoadBalancerHTTPSListenerAttribute"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.SetLoadBalancerHTTPSListenerAttributeResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
