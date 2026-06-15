package ecloudsdkclouddns

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcloudcore"
	"gitlab.ecloud.com/ecloud/ecloudsdkcloudcore/config"
	"gitlab.ecloud.com/ecloud/ecloudsdkclouddns/model"
)

type Client struct {
	APIClient   *ecloudsdkcloudcore.APIClient
	config      *config.Config
	httpRequest *ecloudsdkcloudcore.HttpRequest
}

func NewClient(config *config.Config) *Client {
	client := &Client{}
	client.config = config
	apiClient := ecloudsdkcloudcore.NewAPIClient()
	httpRequest := ecloudsdkcloudcore.NewDefaultHttpRequest()
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
