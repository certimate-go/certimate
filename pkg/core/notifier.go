package core

import (
	"context"
)

// 表示定义消息通知器的抽象类型接口。
type Notifier interface {
	LoggerSetter

	// 发送通知。
	//
	// 入参：
	//   - ctx：上下文。
	//   - subject：通知主题。
	//   - message：通知内容。
	//
	// 出参：
	//   - res：发送结果。
	//   - err: 错误。
	Notify(ctx context.Context, subject, message string) (_res *NotifierNotifyResult, _err error)
}

// 表示通知发送结果的数据结构。
type NotifierNotifyResult struct {
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
