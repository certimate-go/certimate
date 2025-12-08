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
	DisableSDKError *bool
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

func (client *SlbClient) DescribeServerCertificatesWithContext(ctx context.Context, request *alislb.DescribeServerCertificatesRequest, runtime *dara.RuntimeOptions) (_result *alislb.DescribeServerCertificatesResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		query["ResourceGroupId"] = request.ResourceGroupId
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

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("DescribeServerCertificates"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.DescribeServerCertificatesResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}

func (client *SlbClient) UploadServerCertificateWithContext(ctx context.Context, request *alislb.UploadServerCertificateRequest, runtime *dara.RuntimeOptions) (_result *alislb.UploadServerCertificateResponse, _err error) {
	_err = request.Validate()
	if _err != nil {
		return _result, _err
	}

	query := map[string]interface{}{}
	if !dara.IsNil(request.AliCloudCertificateId) {
		query["AliCloudCertificateId"] = request.AliCloudCertificateId
	}

	if !dara.IsNil(request.AliCloudCertificateName) {
		query["AliCloudCertificateName"] = request.AliCloudCertificateName
	}

	if !dara.IsNil(request.AliCloudCertificateRegionId) {
		query["AliCloudCertificateRegionId"] = request.AliCloudCertificateRegionId
	}

	if !dara.IsNil(request.OwnerAccount) {
		query["OwnerAccount"] = request.OwnerAccount
	}

	if !dara.IsNil(request.OwnerId) {
		query["OwnerId"] = request.OwnerId
	}

	if !dara.IsNil(request.PrivateKey) {
		query["PrivateKey"] = request.PrivateKey
	}

	if !dara.IsNil(request.RegionId) {
		query["RegionId"] = request.RegionId
	}

	if !dara.IsNil(request.ResourceGroupId) {
		query["ResourceGroupId"] = request.ResourceGroupId
	}

	if !dara.IsNil(request.ResourceOwnerAccount) {
		query["ResourceOwnerAccount"] = request.ResourceOwnerAccount
	}

	if !dara.IsNil(request.ResourceOwnerId) {
		query["ResourceOwnerId"] = request.ResourceOwnerId
	}

	if !dara.IsNil(request.ServerCertificate) {
		query["ServerCertificate"] = request.ServerCertificate
	}

	if !dara.IsNil(request.ServerCertificateName) {
		query["ServerCertificateName"] = request.ServerCertificateName
	}

	if !dara.IsNil(request.Tag) {
		query["Tag"] = request.Tag
	}

	req := &openapiutil.OpenApiRequest{
		Query: openapiutil.Query(query),
	}
	params := &openapiutil.Params{
		Action:      dara.String("UploadServerCertificate"),
		Version:     dara.String("2014-05-15"),
		Protocol:    dara.String("HTTPS"),
		Pathname:    dara.String("/"),
		Method:      dara.String("POST"),
		AuthType:    dara.String("AK"),
		Style:       dara.String("RPC"),
		ReqBodyType: dara.String("formData"),
		BodyType:    dara.String("json"),
	}
	_result = &alislb.UploadServerCertificateResponse{}
	_body, _err := client.CallApiWithCtx(ctx, params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = dara.Convert(_body, &_result)
	return _result, _err
}
