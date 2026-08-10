package huaweiibmc

import (
	"context"
	"fmt"
	"net/http"
)

type DeleteSessionResponse struct {
	sdkResponseBase
}

func (c *Client) DeleteSession() (*DeleteSessionResponse, error) {
	return c.DeleteSessionWithContext(context.Background())
}

func (c *Client) DeleteSessionWithContext(ctx context.Context) (*DeleteSessionResponse, error) {
	c.tokenMu.Lock()
	if c.token == "" {
		c.tokenMu.Unlock()
		return nil, fmt.Errorf("sdkerr: auth error: session not created")
	}
	if c.tokenLoc == "" {
		c.tokenMu.Unlock()
		return &DeleteSessionResponse{}, nil
	}

	c.tokenMu.Unlock()

	httpreq, err := c.newRequest(http.MethodDelete, c.tokenLoc)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DeleteSessionResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		c.tokenMu.Lock()
		c.token = ""
		c.tokenLoc = ""
		c.tokenMu.Unlock()
	}

	return result, nil
}
