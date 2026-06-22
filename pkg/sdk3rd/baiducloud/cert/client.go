// An extension SDK client for BaiduCloud SSL service.
// Based on github.com/baidubce/bce-sdk-go.
package cert

import (
	"github.com/baidubce/bce-sdk-go/services/cert"
)

type Client struct {
	*cert.Client
}

func NewClient(ak, sk, endPoint string) (*Client, error) {
	client, err := cert.NewClient(ak, sk, endPoint)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}
