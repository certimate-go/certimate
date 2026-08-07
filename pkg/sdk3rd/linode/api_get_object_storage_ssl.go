package linode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type GetObjectStorageSSLResponse struct {
	sdkResponseBase

	SSL bool `json:"ssl"`
}

func (c *Client) GetObjectStorageSSL(regionId string, bucket string) (*GetObjectStorageSSLResponse, error) {
	return c.GetObjectStorageSSLWithContext(context.Background(), regionId, bucket)
}

func (c *Client) GetObjectStorageSSLWithContext(ctx context.Context, regionId string, bucket string) (*GetObjectStorageSSLResponse, error) {
	if regionId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset regionId")
	}
	if bucket == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset bucket")
	}

	path := fmt.Sprintf("/object-storage/buckets/%s/%s/ssl", url.PathEscape(regionId), url.PathEscape(bucket))
	httpreq, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetObjectStorageSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
