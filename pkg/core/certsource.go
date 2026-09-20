package core

import (
	"context"
)

// 表示定义 SSL 证书来源的抽象类型接口。
// 部分服务商通常会提供 SSL 证书托管服务，可供用户集中管理证书，故可从中拉取证书至工作流。
type Certsource interface {
	LoggerSetter

	// 拉取证书。
	//
	// 入参：
	//   - ctx：上下文。
	//   - domain：证书域名。提供商通常以此匹配其托管的证书，若 `certId` 非空则可忽略此参数。
	//   - certId：提供商处的证书 ID。为空时由提供商根据 `domain` 自行匹配。
	//
	// 出参：
	//   - res：拉取结果。
	//   - err: 错误。
	Fetch(ctx context.Context, domain string, certId string) (_res *CertsourceFetchResult, _err error)
}

// 表示 SSL 证书来源拉取结果的数据结构，包含证书、私钥和其他数据。
type CertsourceFetchResult struct {
	CertPEM      string         `json:"certPEM"`
	PrivkeyPEM   string         `json:"privkeyPEM"`
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}
