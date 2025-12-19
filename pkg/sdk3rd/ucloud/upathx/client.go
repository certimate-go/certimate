package upathx

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UPathXClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UPathXClient {
	meta := ucloud.ClientMeta{Product: "PathX"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &UPathXClient{
		client,
	}
}
