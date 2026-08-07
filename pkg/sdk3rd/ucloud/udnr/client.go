// An extension SDK client for UCloud DNR service.
// Based on github.com/ucloud/ucloud-sdk-go.
package udnr

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UDNRClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UDNRClient {
	meta := ucloud.ClientMeta{Product: "UDNR"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	client.GetLogger().SetOutput(io.Discard)
	return &UDNRClient{
		client,
	}
}
