package v2

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

type signer struct {
	apiKey string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://1panel.cn/docs/v2/dev_manual/api_manual/

	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	tokenMd5 := md5.Sum([]byte("1panel" + s.apiKey + timestamp))
	tokenMd5Hex := hex.EncodeToString(tokenMd5[:])

	req.Header.Set("1Panel-Timestamp", timestamp)
	req.Header.Set("1Panel-Token", tokenMd5Hex)

	return nil
}
