package certificate

import (
	common "github.com/certimate-go/certimate/pkg/sdk3rd/wangsu/zz-shared-common"
)

func WithAkSk(ak, sk string) common.OptionsFunc {
	return common.WithAkSk(ak, sk)
}
