package core

import (
	"github.com/go-acme/lego/v5/challenge"
)

// 表示定义 ACME 质询提供者的抽象类型接口。
type ACMEChallenger = challenge.Provider
