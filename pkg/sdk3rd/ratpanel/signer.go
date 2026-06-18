package ratpanel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type signer struct {
	accessTokenId int64
	accessToken   string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://ratpanel.github.io/advanced/api#authentication-mechanism
	// https://acepanel.net/en/advanced/api

	payloadStr := ""
	if req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		payloadStr = string(payloadb)
		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}

	canonicalPath := req.URL.Path
	if !strings.HasPrefix(canonicalPath, "/api") {
		index := strings.Index(canonicalPath, "/api")
		if index != -1 {
			canonicalPath = canonicalPath[index:]
		}
	}

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s",
		req.Method,
		canonicalPath,
		req.URL.Query().Encode(),
		sumSha256(payloadStr),
	)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	stringToSign := fmt.Sprintf("%s\n%s\n%s", "HMAC-SHA256", timestamp, sumSha256(canonicalRequest))

	signature := sumHmacSha256(stringToSign, s.accessToken)

	authorization := fmt.Sprintf("HMAC-SHA256 Credential=%d, Signature=%s", s.accessTokenId, signature)

	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Authorization", authorization)

	return nil
}

func sumSha256(str string) string {
	sum := sha256.Sum256([]byte(str))
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum[:])
	return string(dst)
}

func sumHmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
