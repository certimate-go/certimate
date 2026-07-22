package acm

import (
	common "github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-common"
)

type (
	Options     = common.Options
	OptionsFunc = common.OptionsFunc
)

func WithAkSk(ak, sk string) OptionsFunc {
	return common.WithAkSk(ak, sk)
}

func WithRegion(region string) OptionsFunc {
	return common.WithRegion(region)
}
