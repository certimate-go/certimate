package linode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type DeleteObjectStorageSSLResponse struct {
	sdkResponseBase
}

func (c *Client) DeleteObjectStorageSSL(regionId string, bucket string) (*DeleteObjectStorageSSLResponse, error) {
	return c.DeleteObjectStorageSSLWithContext(context.Background(), regionId, bucket)
}

func (c *Client) DeleteObjectStorageSSLWithContext(ctx context.Context, regionId string, bucket string) (*DeleteObjectStorageSSLResponse, error) {
	if regionId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset regionId")
	}
	if bucket == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset bucket")
	}

	path := fmt.Sprintf("/object-storage/buckets/%s/%s/ssl", url.PathEscape(regionId), url.PathEscape(bucket))
	httpreq, err := c.newRequest(http.MethodDelete, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DeleteObjectStorageSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
