package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type ULBClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *ULBClient {
	meta := ucloud.ClientMeta{Product: "ULB"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &ULBClient{
		client,
	}
}
