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
	type result struct {
		res *RevokeCertificateResponse
		err error
	}

	done := make(chan result, 1)

	go func() {
		res, err := c.sendRevokeCertificateRequest(request)
		done <- result{res, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		return r.res, r.err
	}
}

func (c *ACMEClient) sendRevokeCertificateRequest(request *RevokeCertificateRequest) (*RevokeCertificateResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("the request is nil")
	}

	err := c.client.Certificate.Revoke([]byte(request.Certificate))
	if err != nil {
		return nil, err
	}

	return &RevokeCertificateResponse{}, nil
}
