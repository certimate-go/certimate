package dokploy

import (
	"context"
	"net/http"
)

type UserGetRequest struct{}

type UserGetResponse struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organizationId"`
	UserId         string `json:"userId"`
	Role           string `json:"role"`
	CreatedAt      string `json:"createdAt"`
	TeamId         string `json:"teamId,omitempty"`
	IsDefault      bool   `json:"isDefault"`
	User           *struct {
		Id            string `json:"id"`
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"emailVerified"`
		Role          string `json:"role"`
		CreatedAt     string `json:"createdAt"`
	} `json:"user,omitempty"`
}

func (c *Client) UserGet(req *UserGetRequest) (*UserGetResponse, error) {
	return c.UserGetWithContext(context.Background(), req)
}

func (c *Client) UserGetWithContext(ctx context.Context, req *UserGetRequest) (*UserGetResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/user.get")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &UserGetResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
