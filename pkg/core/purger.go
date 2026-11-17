package core

import (
	"context"
	"time"
)

// 表示定义 SSL 证书清除器的抽象类型接口。
type Purger interface {
	LoggerSetter

	// 清除过期证书。
	//
	// 入参：
	//   - ctx: 上下文。
	//   - expiry: 过期时间。
	//
	// 出参：
	//   - res: 清除结果。
	//   - err: 错误。
	Purge(ctx context.Context, expiry time.Duration) (_res *PurgerPurgeResult, _err error)
}

// 表示 SSL 证书清除结果的数据结构。
type PurgerPurgeResult struct {
	Count        int            `json:"count"`
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
