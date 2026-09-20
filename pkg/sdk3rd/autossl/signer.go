package autossl

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type signer struct {
	accessKey string
	secretKey string

	// 可注入的时钟，便于测试。
	now func() time.Time
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://www.auto-ssl.cn/docs/OpenAPI.html#签名算法
	//
	// 签名原文为「请求方式\n请求路径\n毫秒级时间戳」经 HMAC-SHA256 后以 Base64 编码。
	// 注意：官方示例中请求路径不含查询参数，故此处取 `req.URL.Path`。

	if s.now == nil {
		s.now = time.Now
	}

	timestamp := strconv.FormatInt(s.now().UnixMilli(), 10)

	nonceb := make([]byte, 16)
	if _, err := rand.Read(nonceb); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceb)

	stringToSign := strings.Join([]string{req.Method, req.URL.Path, timestamp}, "\n")

	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-AUTOSSL-API-KEY", s.accessKey)
	req.Header.Set("X-AUTOSSL-TIMESTAMP", timestamp)
	req.Header.Set("X-AUTOSSL-NONCE", nonce)
	req.Header.Set("X-AUTOSSL-SIGNATURE", signature)

	return nil
}
