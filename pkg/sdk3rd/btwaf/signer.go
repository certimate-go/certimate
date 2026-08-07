package btwaf

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type signer struct {
	apiKey string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://github.com/aaPanel/aaWAF/blob/main/API.md

	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	keyMd5 := md5.Sum([]byte(s.apiKey))
	keyMd5Hex := strings.ToLower(hex.EncodeToString(keyMd5[:]))

	signMd5 := md5.Sum([]byte(timestamp + keyMd5Hex))
	signMd5Hex := strings.ToLower(hex.EncodeToString(signMd5[:]))

	req.Header.Set("waf_request_time", timestamp)
	req.Header.Set("waf_request_token", signMd5Hex)

	return nil
}
