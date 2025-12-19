package uewaf

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UEWAFClient struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UEWAFClient {
	meta := ucloud.ClientMeta{Product: "UEWAF"}
	client := ucloud.NewClientWithMeta(config, credential, meta)
	return &UEWAFClient{
		client,
	}
}
