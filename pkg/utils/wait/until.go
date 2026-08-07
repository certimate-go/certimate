package wait

import (
	"context"
	"time"
)

// 等待直到条件满足。
//
// 入参：
//   - condition: 条件函数，接收尝试次数作为参数，返回是否满足条件和错误。
//   - interval: 执行条件函数的间隔时间。
//
// 出参：
//   - ret: 是否满足条件。
//   - err: 错误。
func Until(condition func(index int) (bool, error), interval time.Duration) (bool, error) {
	conditionWithContext := func(_ context.Context, index int) (bool, error) {
		return condition(index)
	}
	return UntilWithContext(context.Background(), conditionWithContext, interval)
}

// 等待直到条件满足，支持传入 context.Context 上下文。
//
// 入参：
//   - ctx: 上下文。
//   - condition: 条件函数，接收上下文和尝试次数作为参数，返回是否满足条件和错误。
//   - interval: 执行条件函数的间隔时间。
//
// 出参：
//   - ret: 是否满足条件。
//   - err: 错误。
func UntilWithContext(ctx context.Context, condition func(ctx context.Context, index int) (bool, error), interval time.Duration) (bool, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()

		case <-ticker.C:
			attempt++
			ret, err := condition(ctx, attempt)
			if ret || err != nil {
				return ret, err
			}
		}
	}
}

// 等待直到条件满足或超时。
//
// 入参：
//   - condition: 条件函数，接收尝试次数作为参数，返回是否满足条件和错误。
//   - timeout: 超时时间。
//   - interval: 执行条件函数的间隔时间。
//
// 出参：
//   - ret: 是否满足条件。
//   - err: 错误。
func UntilTimeout(condition func(index int) (bool, error), timeout time.Duration, interval time.Duration) (bool, error) {
	conditionWithContext := func(_ context.Context, index int) (bool, error) {
		return condition(index)
	}
	return UntilTimeoutWithContext(context.Background(), conditionWithContext, timeout, interval)
}

// 等待直到条件满足或超时，支持传入 context.Context 上下文。
//
// 入参：
//   - ctx: 上下文。
//   - condition: 条件函数，接收上下文和尝试次数作为参数，返回是否满足条件和错误。
//   - timeout: 超时时间。
//   - interval: 执行条件函数的间隔时间。
//
// 出参：
//   - ret: 是否满足条件。
//   - err: 错误。
func UntilTimeoutWithContext(ctx context.Context, condition func(ctx context.Context, index int) (bool, error), timeout time.Duration, interval time.Duration) (bool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	attempt := 0
	for {
		select {
		case <-ctxWithTimeout.Done():
			return false, ctx.Err()

		case <-ticker.C:
			attempt++
			ret, err := condition(ctxWithTimeout, attempt)
			if ret || err != nil {
				return ret, err
			}
		}
	}
}
