package btpanel

import (
	"context"
	"net/http"
)

type FilesUploadRequest struct {
	Path  *string `json:"path,omitempty"`
	Name  *string `json:"filename,omitempty"`
	Start *int32  `json:"start,omitempty"`
	Size  *int32  `json:"size,omitempty"`
	Blob  []byte  `json:"-" form:"blob"`
	Force *bool   `json:"force,omitempty"`
}

type FilesUploadResponse struct {
	apiResponseBase
}

func (c *Client) FilesUpload(req *FilesUploadRequest) (*FilesUploadResponse, error) {
	return c.FilesUploadWithContext(context.Background(), req)
}

func (c *Client) FilesUploadWithContext(ctx context.Context, req *FilesUploadRequest) (*FilesUploadResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/files/upload", req, true)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &FilesUploadResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
