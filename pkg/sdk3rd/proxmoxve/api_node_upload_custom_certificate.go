package proxmoxve

import (
	"context"
	"fmt"
	"net/http"
)

type NodeUploadCustomCertificateRequest CustomCertificate

type NodeUploadCustomCertificateResponse struct {
	Data *struct {
		FileName    string   `json:"filename,omitempty"`
		Fingerprint string   `json:"fingerprint,omitempty"`
		Subject     string   `json:"subject,omitempty"`
		Issuer      string   `json:"issuer,omitempty"`
		SAN         []string `json:"san,omitempty"`
		NotAfter    int64    `json:"notafter,omitempty"`
		NotBefore   int64    `json:"notbefore,omitempty"`
	} `json:"data,omitempty"`
}

func (c *Client) NodeUploadCustomCertificate(node string, req *NodeUploadCustomCertificateRequest) (*NodeUploadCustomCertificateResponse, error) {
	return c.NodeUploadCustomCertificateWithContext(context.Background(), node, req)
}

func (c *Client) NodeUploadCustomCertificateWithContext(ctx context.Context, node string, req *NodeUploadCustomCertificateRequest) (*NodeUploadCustomCertificateResponse, error) {
	if node == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset node")
	}

	path := fmt.Sprintf("/nodes/%s/certificates/custom", node)
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &NodeUploadCustomCertificateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
