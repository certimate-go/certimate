package asiaispcdn

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type signer struct {
	accessKeyId     string
	accessKeySecret string
}

func (s *signer) Sign(req *http.Request) error {
	method := strings.ToUpper(req.Method)

	pathStr := "/"
	queryStr := ""
	if req.URL != nil {
		pathStr = req.URL.Path
		queryStr = req.URL.Query().Encode()
	}

	payloadStr := ""
	if method != http.MethodGet && method != http.MethodDelete && req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		payloadStr = string(payloadb)
		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}

	nonce := 10000000 + rand.Intn(90000000)
	timestamp := time.Now().UnixMilli()

	stringToSign := ""
	if method == http.MethodGet || method == http.MethodDelete {
		stringToSign = fmt.Sprintf("accessKeySecret=%s&method=%s&nonce=%d&queryString=%s&timestamp=%d&uri=%s", s.accessKeySecret, method, nonce, queryStr, timestamp, pathStr)
	} else {
		stringToSign = fmt.Sprintf("accessKeySecret=%s&body=%s&method=%s&nonce=%d&queryString=%s&timestamp=%d&uri=%s", s.accessKeySecret, payloadStr, method, nonce, queryStr, timestamp, pathStr)
	}

	h := hmac.New(sha1.New, []byte(s.accessKeySecret))
	h.Write([]byte(stringToSign))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	req.Header.Set("accessKeyId", s.accessKeyId)
	req.Header.Set("nonce", fmt.Sprintf("%d", nonce))
	req.Header.Set("timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("signature", signature)

	return nil
}
