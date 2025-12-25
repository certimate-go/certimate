package internal

import (
	"context"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	aliwaf "github.com/alibabacloud-go/waf-openapi-20211001/v7/client"
)

// This is a partial copy of https://github.com/alibabacloud-go/waf-openapi-20211001/blob/master/client/client_context_func.go
// to lightweight the vendor packages in the built binary.
type WafClient struct {
	openapi.Client
	DisableSDKError *bool
}

func NewWafClient(config *openapiutil.Config) (*WafClient, error) {
	client := new(WafClient)
	err := client.Init(config)
	return client, err
}

func (client *WafClient) Init(config *openapiutil.Config) (_err error) {
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

func (client *WafClient) DescribeDefaultHttpsWithContext(ctx context.Context, request *aliwaf.DescribeDefaultHttpsRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.DescribeDefaultHttpsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeDefaultHttps"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.DescribeDefaultHttpsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) DescribeDomainDetailWithContext(ctx context.Context, request *aliwaf.DescribeDomainDetailRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.DescribeDomainDetailResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.Domain) {
		query["Domain"] = request.Domain
	}

	if !dara.IsNil(request.DomainId) {
		query["DomainId"] = request.DomainId
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeDomainDetail"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.DescribeDomainDetailResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) DescribeProductInstancesWithContext(ctx context.Context, request *aliwaf.DescribeProductInstancesRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.DescribeProductInstancesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.OwnerUserId) {
		query["OwnerUserId"] = request.OwnerUserId
	}

	if !dara.IsNil(request.PageNumber) {
		query["PageNumber"] = request.PageNumber
	}

	if !dara.IsNil(request.PageSize) {
		query["PageSize"] = request.PageSize
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceInstanceAccessStatus) {
		query["ResourceInstanceAccessStatus"] = request.ResourceInstanceAccessStatus
	}

	if !dara.IsNil(request.ResourceInstanceId) {
		query["ResourceInstanceId"] = request.ResourceInstanceId
	}

	if !dara.IsNil(request.ResourceInstanceIp) {
		query["ResourceInstanceIp"] = request.ResourceInstanceIp
	}

	if !dara.IsNil(request.ResourceInstanceName) {
		query["ResourceInstanceName"] = request.ResourceInstanceName
	}

	if !dara.IsNil(request.ResourceIp) {
		query["ResourceIp"] = request.ResourceIp
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	if !dara.IsNil(request.ResourceName) {
		query["ResourceName"] = request.ResourceName
	}

	if !dara.IsNil(request.ResourceProduct) {
		query["ResourceProduct"] = request.ResourceProduct
	}

	if !dara.IsNil(request.ResourceRegionId) {
		query["ResourceRegionId"] = request.ResourceRegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeProductInstances"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.DescribeProductInstancesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) DescribeResourceInstanceCertsWithContext(ctx context.Context, request *aliwaf.DescribeResourceInstanceCertsRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.DescribeResourceInstanceCertsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.PageNumber) {
		query["PageNumber"] = request.PageNumber
	}

	if !dara.IsNil(request.PageSize) {
		query["PageSize"] = request.PageSize
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceInstanceId) {
		query["ResourceInstanceId"] = request.ResourceInstanceId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeResourceInstanceCerts"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.DescribeResourceInstanceCertsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) ModifyCloudResourceWithContext(ctx context.Context, tmpReq *aliwaf.ModifyCloudResourceRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.ModifyCloudResourceResponse, _err error) {
	_err = tmpReq.Validate()
	if _err != nil {
		return _result, _err
	}

	request := &aliwaf.ModifyCloudResourceShrinkRequest{}
	openapiutil.Convert(tmpReq, request)

	if !dara.IsNil(tmpReq.Listen) {
		request.ListenShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.Listen, dara.String("Listen"), dara.String("json"))
	}

	if !dara.IsNil(tmpReq.Redirect) {
		request.RedirectShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.Redirect, dara.String("Redirect"), dara.String("json"))
	}

	query := map[string]interface{}{}

	if !dara.IsNil(request.CloudResourceId) {
		query["CloudResourceId"] = request.CloudResourceId
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.ListenShrink) {
		query["Listen"] = request.ListenShrink
	}

	if !dara.IsNil(request.RedirectShrink) {
		query["Redirect"] = request.RedirectShrink
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ModifyCloudResource"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.ModifyCloudResourceResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) ModifyDefaultHttpsWithContext(ctx context.Context, request *aliwaf.ModifyDefaultHttpsRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.ModifyDefaultHttpsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}

	if !dara.IsNil(request.CertId) {
		query["CertId"] = request.CertId
	}

	if !dara.IsNil(request.CipherSuite) {
		query["CipherSuite"] = request.CipherSuite
	}

	if !dara.IsNil(request.CustomCiphers) {
		query["CustomCiphers"] = request.CustomCiphers
	}

	if !dara.IsNil(request.EnableTLSv3) {
		query["EnableTLSv3"] = request.EnableTLSv3
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	if !dara.IsNil(request.TLSVersion) {
		query["TLSVersion"] = request.TLSVersion
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ModifyDefaultHttps"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.ModifyDefaultHttpsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) ModifyDomainWithContext(ctx context.Context, tmpReq *aliwaf.ModifyDomainRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.ModifyDomainResponse, _err error) {
	_err = tmpReq.Validate()
	if _err != nil {
		return _result, _err
	}

	request := &aliwaf.ModifyDomainShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !dara.IsNil(tmpReq.Listen) {
		request.ListenShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.Listen, dara.String("Listen"), dara.String("json"))
	}

	if !dara.IsNil(tmpReq.Redirect) {
		request.RedirectShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.Redirect, dara.String("Redirect"), dara.String("json"))
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.AccessType) {
		query["AccessType"] = request.AccessType
	}

	if !dara.IsNil(request.Domain) {
		query["Domain"] = request.Domain
	}

	if !dara.IsNil(request.DomainId) {
		query["DomainId"] = request.DomainId
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.ListenShrink) {
		query["Listen"] = request.ListenShrink
	}

	if !dara.IsNil(request.RedirectShrink) {
		query["Redirect"] = request.RedirectShrink
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ModifyDomain"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.ModifyDomainResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) CreateCertsWithContext(ctx context.Context, request *aliwaf.CreateCertsRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.CreateCertsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.CertContent) {
		query["CertContent"] = request.CertContent
	}

	if !dara.IsNil(request.CertKey) {
		query["CertKey"] = request.CertKey
	}

	if !dara.IsNil(request.CertName) {
		query["CertName"] = request.CertName
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("CreateCerts"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.CreateCertsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) DescribeCloudResourceAccessPortDetailsWithContext(ctx context.Context, request *aliwaf.DescribeCloudResourceAccessPortDetailsRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.DescribeCloudResourceAccessPortDetailsResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.PageNumber) {
		query["PageNumber"] = request.PageNumber
	}

	if !dara.IsNil(request.PageSize) {
		query["PageSize"] = request.PageSize
	}

	if !dara.IsNil(request.Port) {
		query["Port"] = request.Port
	}

	if !dara.IsNil(request.Protocol) {
		query["Protocol"] = request.Protocol
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceInstanceId) {
		query["ResourceInstanceId"] = request.ResourceInstanceId
	}

	if !dara.IsNil(request.ResourceManagerResourceGroupId) {
		query["ResourceManagerResourceGroupId"] = request.ResourceManagerResourceGroupId
	}

	if !dara.IsNil(request.ResourceProduct) {
		query["ResourceProduct"] = request.ResourceProduct
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeCloudResourceAccessPortDetails"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.DescribeCloudResourceAccessPortDetailsResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *WafClient) ModifyCloudResourceCertWithContext(ctx context.Context, request *aliwaf.ModifyCloudResourceCertRequest, runtime *dara.RuntimeOptions) (_result *aliwaf.ModifyCloudResourceCertResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !dara.IsNil(request.Certificates) {
		query["Certificates"] = request.Certificates
	}

	if !dara.IsNil(request.CloudResourceId) {
		query["CloudResourceId"] = request.CloudResourceId
	}

	if !dara.IsNil(request.InstanceId) {
		query["InstanceId"] = request.InstanceId
	}

	if !dara.IsNil(request.Port) {
		query["Port"] = request.Port
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceInstanceId) {
		query["ResourceInstanceId"] = request.ResourceInstanceId
	}

	if !dara.IsNil(request.ResourceProduct) {
		query["ResourceProduct"] = request.ResourceProduct
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("ModifyCloudResourceCert"),
		Version:     dara.String("2021-10-01"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &aliwaf.ModifyCloudResourceCertResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
