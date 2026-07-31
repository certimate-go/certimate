// An extension SDK client for UCloud PathX service.
// Based on github.com/ucloud/ucloud-sdk-go.
package upathx

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UPathXClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UPathXClient {
	meta := ucloud.ClientMeta{Product: "PathX"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	client.GetLogger().SetOutput(io.Discard)
	return &UPathXClient{
		client,
	}
}
