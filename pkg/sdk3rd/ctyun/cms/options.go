package cms

import (
	common "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/zz-shared-common"
)

func WithAkSk(ak, sk string) common.OptionsFunc {
	return common.WithAkSk(ak, sk)
}
