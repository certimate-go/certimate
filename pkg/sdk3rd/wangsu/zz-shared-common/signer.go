package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type signer struct {
	accessKey string
	secretKey string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://www.wangsu.com/document/openapi/api-authentication

	method := strings.ToUpper(req.Method)

	path := "/"
	if req.URL != nil {
		path = req.URL.Path
	}

	queryStr := ""
	if method != http.MethodPost && req.URL != nil {
		queryStr = req.URL.RawQuery

		s, err := url.QueryUnescape(queryStr)
		if err != nil {
			return err
		}

		queryStr = s
	}

	canonicalHeaders := "" +
		"content-type:" + strings.TrimSpace(strings.ToLower(req.Header.Get("Content-Type"))) + "\n" +
		"host:" + strings.TrimSpace(strings.ToLower(req.Host)) + "\n"
	signedHeaders := "content-type;host"

	payloadStr := ""
	if method != http.MethodGet && req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		payloadStr = string(payloadb)
		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}
	payloadHash := sha256.Sum256([]byte(payloadStr))
	payloadHashHex := strings.ToLower(hex.EncodeToString(payloadHash[:]))

	nowUtc := time.Now().UTC()
	timestampStr := req.Header.Get("X-CNC-Timestamp")
	if timestampStr == "" {
		timestampStr = fmt.Sprintf("%d", nowUtc.Unix())
	} else {
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return err
		}
		nowUtc = time.Unix(timestamp, 0).UTC()
	}

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, path, queryStr, canonicalHeaders, signedHeaders, payloadHashHex)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

	const signAlgorithmHeader = "CNC-HMAC-SHA256"
	stringToSign := fmt.Sprintf("%s\n%s\n%s", signAlgorithmHeader, timestampStr, canonicalRequestHashHex)

	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(stringToSign))
	signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

	authorization := fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s", signAlgorithmHeader, s.accessKey, signedHeaders, signature)

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Date", nowUtc.Format(http.TimeFormat))
	req.Header.Set("X-CNC-Auth-Method", "AKSK")
	req.Header.Set("X-CNC-AccessKey", s.accessKey)
	req.Header.Set("X-CNC-Timestamp", timestampStr)

	return nil
}
