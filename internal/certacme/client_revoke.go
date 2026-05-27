package certacme

import (
	"context"
	"fmt"
)

type RevokeCertificateRequest struct {
	Certificate string
}

type RevokeCertificateResponse struct{}

func (c *ACMEClient) RevokeCertificate(ctx context.Context, request *RevokeCertificateRequest) (*RevokeCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	err := c.client.Certificate.Revoke(ctx, []byte(request.Certificate))
	if err != nil {
		return nil, err
	}

	return &RevokeCertificateResponse{}, nil
}
