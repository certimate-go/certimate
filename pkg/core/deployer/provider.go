package deployer

import (
	"context"
	"log/slog"
)

// 表示定义 SSL 证书部署器的抽象类型接口。
type Provider interface {
	// 设置日志记录器。
	//
	// 入参：
	//   - logger：日志记录器实例。
	SetLogger(logger *slog.Logger)

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
	Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (_res *DeployResult, _err error)
}

// 表示 SSL 证书部署结果的数据结构。
type DeployResult struct {
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
