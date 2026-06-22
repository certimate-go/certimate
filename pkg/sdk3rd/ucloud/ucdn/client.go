// An extension SDK client for UCloud CDN service.
// Based on github.com/ucloud/ucloud-sdk-go.
package ucdn

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UCDNClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UCDNClient {
	meta := ucloud.ClientMeta{Product: "UCDN"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &UCDNClient{
		client,
	}
}
