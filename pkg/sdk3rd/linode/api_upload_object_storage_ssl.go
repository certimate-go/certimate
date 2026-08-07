package linode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UploadObjectStorageSSLRequest struct {
	sdkResponseBase

	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

type UploadObjectStorageSSLResponse struct {
	sdkResponseBase

	SSL bool `json:"ssl"`
}

func (c *Client) UploadObjectStorageSSL(regionId string, bucket string, req *UploadObjectStorageSSLRequest) (*UploadObjectStorageSSLResponse, error) {
	return c.UploadObjectStorageSSLWithContext(context.Background(), regionId, bucket, req)
}

func (c *Client) UploadObjectStorageSSLWithContext(ctx context.Context, regionId string, bucket string, req *UploadObjectStorageSSLRequest) (*UploadObjectStorageSSLResponse, error) {
	if regionId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset regionId")
	}
	if bucket == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset bucket")
	}

	path := fmt.Sprintf("/object-storage/buckets/%s/%s/ssl", url.PathEscape(regionId), url.PathEscape(bucket))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UploadObjectStorageSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
