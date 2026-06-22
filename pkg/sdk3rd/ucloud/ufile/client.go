// An extension SDK client for UCloud OS service.
// Based on github.com/ucloud/ucloud-sdk-go.
package ufile

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UFileClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UFileClient {
	meta := ucloud.ClientMeta{Product: "UFile"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &UFileClient{
		client,
	}
}
