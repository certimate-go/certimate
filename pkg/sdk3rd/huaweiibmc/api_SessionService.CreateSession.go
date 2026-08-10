package huaweiibmc

import (
	"context"
	"fmt"
	"net/http"
)

type CreateSessionResponse struct {
	sdkResponseBase

	XToken    string `json:"-"`
	XLocation string `json:"-"`
}

func (c *Client) CreateSession() (*CreateSessionResponse, error) {
	return c.CreateSessionWithContext(context.Background())
}

func (c *Client) CreateSessionWithContext(ctx context.Context) (*CreateSessionResponse, error) {
	c.tokenMu.Lock()
	if c.token != "" {
		c.tokenMu.Unlock()
		return nil, fmt.Errorf("sdkerr: auth error: session already created")
	}

	c.tokenMu.Unlock()

	httpreq, err := c.newRequest(http.MethodPost, "/redfish/v1/SessionService/Sessions")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(map[string]string{
			"UserName": c.username,
			"Password": c.password,
		})
		httpreq.SetContext(ctx)
	}

	result := &CreateSessionResponse{}
	if httpresp, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		token := httpresp.Header().Get("X-Auth-Token")
		location := httpresp.Header().Get("Location")
		result.XToken = token
		result.XLocation = c.resolveLocation(location)

		if result.XToken == "" {
			return result, fmt.Errorf("sdkerr: auth error: received empty token")
		}

		c.tokenMu.Lock()
		c.token = result.XToken
		c.tokenLoc = result.XLocation
		c.tokenMu.Unlock()
	}

	return result, nil
}
