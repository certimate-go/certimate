package core

import (
	"context"
)

// 表示定义 SSL 证书管理器的抽象类型接口。
// 云服务商通常会提供 SSL 证书管理服务，可供用户集中管理证书。
type Certmgr interface {
	LoggerSetter

	// 上传证书。
	//
	// 入参：
	//   - ctx：上下文。
	//   - certPEM：证书 PEM 内容。
	//   - privkeyPEM：私钥 PEM 内容。
	//
	// 出参：
	//   - res：上传结果。
	//   - err: 错误。
	Upload(ctx context.Context, certPEM, privkeyPEM string) (_res *CertmgrUploadResult, _err error)

	// 替换证书。
	//
	// 入参：
	//   - ctx：上下文。
	//   - certIdOrName：证书 ID 或名称，即云服务商处的证书标识符。
	//   - certPEM：证书 PEM 内容。
	//   - privkeyPEM：私钥 PEM 内容。
	//
	// 出参：
	//   - res：替换结果。
	//   - err: 错误。
	Replace(ctx context.Context, certIdOrName string, certPEM, privkeyPEM string) (_res *CertmgrReplaceResult, _err error)
}

// 表示 SSL 证书管理替换结果的数据结构。
type CertmgrReplaceResult struct {
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}

// 表示 SSL 证书管理上传结果的数据结构，包含上传后的证书 ID、名称和其他数据。
type CertmgrUploadResult struct {
	CertId       string         `json:"certId,omitempty"`
	CertName     string         `json:"certName,omitempty"`
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
