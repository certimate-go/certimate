package dogecloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

type signer struct {
	accessKey string
	secretKey string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://docs.dogecloud.com/cdn/api-access-token

	path := req.URL.Path
	queryStr := req.URL.Query().Encode()
	if queryStr != "" {
		path += "?" + queryStr
	}

	payloadStr := ""
	if req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		payloadStr = string(payloadb)
		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}

	stringToSign := fmt.Sprintf("%s\n%s", path, payloadStr)

	h := hmac.New(sha1.New, []byte(s.secretKey))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	authorization := fmt.Sprintf("TOKEN %s:%s", s.accessKey, signature)

	req.Header.Set("Authorization", authorization)

	return nil
}
