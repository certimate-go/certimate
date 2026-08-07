package cdnpro

import (
	"context"
	"net/http"
)

type CreateDeploymentTaskRequest struct {
	Name    *string                 `json:"name,omitempty"`
	Target  *string                 `json:"target,omitempty"`
	Actions *[]DeploymentTaskAction `json:"actions,omitempty"`
	Webhook *string                 `json:"webhook,omitempty"`
}

type CreateDeploymentTaskResponse struct {
	sdkResponseBase

	DeploymentTaskLocation string `json:"location,omitempty"`
}

func (c *Client) CreateDeploymentTask(req *CreateDeploymentTaskRequest) (*CreateDeploymentTaskResponse, error) {
	return c.CreateDeploymentTaskWithContext(context.Background(), req)
}

func (c *Client) CreateDeploymentTaskWithContext(ctx context.Context, req *CreateDeploymentTaskRequest) (*CreateDeploymentTaskResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/cdn/deploymentTasks")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CreateDeploymentTaskResponse{}
	if httpresp, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		result.DeploymentTaskLocation = httpresp.Header().Get("Location")
	}

	return result, nil
}
