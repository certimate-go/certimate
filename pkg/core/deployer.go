package core

import (
	"context"
)

// 表示定义 SSL 证书部署器的抽象类型接口。
type Deployer interface {
	LoggerSetter

	// 部署证书。
	//
	// 入参：
	//   - ctx：上下文。
	//   - certPEM：证书 PEM 内容。
	//   - privkeyPEM：私钥 PEM 内容。
	//
	// 出参：
	//   - res：部署结果。
	//   - err: 错误。
	Deploy(ctx context.Context, certPEM, privkeyPEM string) (_res *DeployerDeployResult, _err error)
}

// 表示 SSL 证书部署结果的数据结构。
type DeployerDeployResult struct {
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
