package wait

import (
	"context"
	"time"
)

// 等待，直到指定时间。
//
// 入参：
//   - wait: 等待时间。
//
// 出参：
//   - err: 错误。
func Delay(wait time.Duration) error {
	return DelayWithContext(context.Background(), wait)
}

// 等待，直到指定时间，或上下文被取消。
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

// 等待，直到条件满足或超时。
//
// 入参：
//   - condition: 条件函数，接收尝试次数作为参数，返回是否满足条件和错误。
//   - timeout: 超时时间。
//   - interval: 执行条件函数的间隔时间。
//
// 出参：
//   - ret: 是否满足条件。
//   - err: 错误。
func For(condition func(int) (bool, error), timeout time.Duration, interval time.Duration) (bool, error) {
	conditionWithContext := func(_ context.Context, i int) (bool, error) {
		return condition(i)
	}
	return ForWithContext(context.Background(), conditionWithContext, timeout, interval)
}

// 等待，直到条件满足或超时，或上下文被取消。
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
func ForWithContext(ctx context.Context, condition func(context.Context, int) (bool, error), timeout time.Duration, interval time.Duration) (bool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	attempt := 0
	for {
		select {
		case <-ctxWithTimeout.Done():
			return false, ctxWithTimeout.Err()

		case <-ticker.C:
			attempt++
			ret, err := condition(ctxWithTimeout, attempt)
			if ret || err != nil {
				return ret, err
			}
		}
	}
}
