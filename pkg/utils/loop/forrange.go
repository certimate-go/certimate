package loop

import (
	"context"
	"errors"
)

// 遍历集合并执行迭代函数。
//
// 入参：
//   - collection: 集合。
//   - iter: 迭代函数，接收集合中的元素和索引作为参数，返回错误。
//
// 出参：
//   - err: 错误。
func ForRange[T any](collection []T, iter func(item T, index int) error) error {
	iterWithContext := func(_ context.Context, item T, index int) error {
		return iter(item, index)
	}
	return ForRangeWithContext(context.Background(), collection, iterWithContext)
}

// 遍历集合并执行迭代函数，支持传入 context.Context 上下文。
//
// 入参：
//   - ctx: 上下文。
//   - collection: 集合。
//   - iter: 迭代函数，接收上下文、集合中的元素和索引作为参数，返回错误。
//
// 出参：
//   - err: 错误。
func ForRangeWithContext[T any](ctx context.Context, collection []T, iter func(ctx context.Context, item T, index int) error) error {
	for i, item := range collection {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			if err := iter(ctx, item, i); err != nil {
				return err
			}
		}
	}

	return nil
}

// 与 [ForRange] 类似，但在迭代时遇到错误不会中止、并在收集所有错误后返回。
//
// 入参：
//   - collection: 集合。
//   - iter: 迭代函数，接收集合中的元素和索引作为参数，返回错误。
//
// 出参：
//   - err: 错误。
func ForRangeAll[T any](collection []T, iter func(item T, index int) error) error {
	iterWithContext := func(_ context.Context, item T, index int) error {
		return iter(item, index)
	}
	return ForRangeAllWithContext(context.Background(), collection, iterWithContext)
}

// 与 [ForRangeAllWithContext] 类似，但在迭代时遇到错误不会中止、并在收集所有错误后返回。
//
// 入参：
//   - ctx: 上下文。
//   - collection: 集合。
//   - iter: 迭代函数，接收上下文、集合中的元素和索引作为参数，返回错误。
//
// 出参：
//   - err: 错误。
func ForRangeAllWithContext[T any](ctx context.Context, collection []T, iter func(ctx context.Context, item T, index int) error) error {
	var errs []error

	for i, item := range collection {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			if err := iter(ctx, item, i); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
