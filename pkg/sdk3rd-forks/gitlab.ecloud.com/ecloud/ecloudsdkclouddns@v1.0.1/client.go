// @Title  Golang SDK Client
// @Description  This code is auto generated
// @Author  Ecloud SDK

package ecloudsdkclouddns

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkclouddns/model"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
)

type Client struct {
	APIClient   *ecloudsdkcore.APIClient
	config      *config.Config
	httpRequest *ecloudsdkcore.HttpRequest
}

func NewClient(config *config.Config) *Client {
	client := &Client{}
	client.config = config
	apiClient := ecloudsdkcore.NewAPIClient()
	httpRequest := ecloudsdkcore.NewDefaultHttpRequest()
	httpRequest.Product = product
	httpRequest.Version = version
	httpRequest.SdkVersion = sdkVersion
	client.httpRequest = httpRequest
	client.APIClient = apiClient
	return client
}

func NewClientByCustomized(config *config.Config, httpRequest *ecloudsdkcore.HttpRequest) *Client {
	client := &Client{}
	client.config = config
	apiClient := ecloudsdkcore.NewAPIClient()
	httpRequest.Product = product
	httpRequest.Version = version
	httpRequest.SdkVersion = sdkVersion
	client.httpRequest = httpRequest
	client.APIClient = apiClient
	return client
}

const (
	product    string = "clouddns"
	version    string = "v1"
	sdkVersion string = "1.0.1"
)

// CreateRecord 新增解析记录
func (c *Client) CreateRecord(request *model.CreateRecordRequest) (*model.CreateRecordResponse, error) {
	c.httpRequest.Action = "createRecord"
	c.httpRequest.Body = request
	returnValue := &model.CreateRecordResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// CreateRecordOpenapi 新增解析记录Openapi
func (c *Client) CreateRecordOpenapi(request *model.CreateRecordOpenapiRequest) (*model.CreateRecordOpenapiResponse, error) {
	c.httpRequest.Action = "createRecordOpenapi"
	c.httpRequest.Body = request
	returnValue := &model.CreateRecordOpenapiResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// DeleteRecord 删除解析记录
func (c *Client) DeleteRecord(request *model.DeleteRecordRequest) (*model.DeleteRecordResponse, error) {
	c.httpRequest.Action = "deleteRecord"
	c.httpRequest.Body = request
	returnValue := &model.DeleteRecordResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// DeleteRecordOpenapi 删除解析记录Openapi
func (c *Client) DeleteRecordOpenapi(request *model.DeleteRecordOpenapiRequest) (*model.DeleteRecordOpenapiResponse, error) {
	c.httpRequest.Action = "deleteRecordOpenapi"
	c.httpRequest.Body = request
	returnValue := &model.DeleteRecordOpenapiResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// ListRecord 查询解析记录
func (c *Client) ListRecord(request *model.ListRecordRequest) (*model.ListRecordResponse, error) {
	c.httpRequest.Action = "listRecord"
	c.httpRequest.Body = request
	returnValue := &model.ListRecordResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// ListRecordOpenapi 查询解析记录Openapi
func (c *Client) ListRecordOpenapi(request *model.ListRecordOpenapiRequest) (*model.ListRecordOpenapiResponse, error) {
	c.httpRequest.Action = "listRecordOpenapi"
	c.httpRequest.Body = request
	returnValue := &model.ListRecordOpenapiResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// ModifyRecord 修改解析记录
func (c *Client) ModifyRecord(request *model.ModifyRecordRequest) (*model.ModifyRecordResponse, error) {
	c.httpRequest.Action = "modifyRecord"
	c.httpRequest.Body = request
	returnValue := &model.ModifyRecordResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

// ModifyRecordOpenapi 修改解析记录Openapi
func (c *Client) ModifyRecordOpenapi(request *model.ModifyRecordOpenapiRequest) (*model.ModifyRecordOpenapiResponse, error) {
	c.httpRequest.Action = "modifyRecordOpenapi"
	c.httpRequest.Body = request
	returnValue := &model.ModifyRecordOpenapiResponse{}
	if _, err := c.APIClient.Excute(c.httpRequest, c.config, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}
