package wait

import (
	"context"
	"time"
)

// 等待一段时间。
//
// 入参：
//   - wait: 等待时间。
//
// 出参：
//   - err: 错误。
func Delay(wait time.Duration) error {
	return DelayWithContext(context.Background(), wait)
}

// 等待一段时间，支持传入 context.Context 上下文。
//
// 入参：
//   - ctx: 上下文。
//   - wait: 等待时间。
//
// 出参：
//   - err: 错误。
func DelayWithContext(ctx context.Context, wait time.Duration) error {
	ticker := time.NewTimer(wait)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ticker.C:
		return nil
	}
}
