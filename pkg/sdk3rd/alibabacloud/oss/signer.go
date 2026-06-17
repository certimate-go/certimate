package oss

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type signer struct {
	accessKeyId     string
	accessKeySecret string
	region          string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://help.aliyun.com/zh/oss/developer-reference/recommend-to-use-signature-version-4
	// https://www.alibabacloud.com/help/en/oss/developer-reference/recommend-to-use-signature-version-4

	method := strings.ToUpper(req.Method)

	nowUtc := time.Now().UTC()
	headerDateStr := nowUtc.Format(http.TimeFormat)
	requestDateStr := nowUtc.Format("20060102T150405Z")
	signDateStr := nowUtc.Format("20060102")

	requestResStr := req.Header.Get("X-API-Resource")
	req.Header.Del("X-API-Resource")

	canonicalUrl := escapePath(req.URL.Path)
	if canonicalUrl == "" {
		canonicalUrl = "/"
	}
	if canonicalUrl == "/" && requestResStr != "" {
		canonicalUrl = escapePath(requestResStr)
	}

	canonicalQueryStr := ""
	if len(req.URL.Query()) > 0 {
		query := req.URL.Query()

		keys := make([]string, 0, len(query))
		for key := range query {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for i, key := range keys {
			if i > 0 {
				canonicalQueryStr += "&"
			}

			value := query.Get(key)
			if value == "" {
				canonicalQueryStr += escapeQuery(key)
			} else {
				canonicalQueryStr += escapeQuery(key) + "=" + escapeQuery(value)
			}
		}
	}

	canonicalHeaders := ""
	additionalHeaders := ""
	if len(req.Header) > 0 {
		if req.Header.Get("X-OSS-Date") == "" {
			req.Header.Set("X-OSS-Date", requestDateStr)
		}
		if req.Header.Get("X-OSS-Content-SHA256") == "" {
			req.Header.Set("X-OSS-Content-SHA256", "UNSIGNED-PAYLOAD")
		}

		keys := make([]string, 0, len(req.Header))
		for key := range req.Header {
			key = strings.ToLower(key)
			if strings.HasPrefix(key, "x-oss-") {
				keys = append(keys, key)
			}
			if key == "content-type" || key == "content-md5" {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)

		for i, key := range keys {
			if i > 0 {
				canonicalHeaders += "\n"
			}

			value := strings.TrimSpace(req.Header.Get(key))
			canonicalHeaders += key + ":" + value
		}

		canonicalHeaders += "\n"
	}

	hashedPayload := req.Header.Get("X-OSS-Content-SHA256")

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, canonicalUrl, canonicalQueryStr, canonicalHeaders, additionalHeaders, hashedPayload)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

	const signAlgorithmHeader = "OSS4-HMAC-SHA256"
	scope := fmt.Sprintf("%s/%s/oss/aliyun_v4_request", signDateStr, s.region)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", signAlgorithmHeader, requestDateStr, scope, canonicalRequestHashHex)

	var h hash.Hash
	h = hmac.New(sha256.New, []byte("aliyun_v4"+s.accessKeySecret))
	h.Write([]byte(signDateStr))
	kDate := h.Sum(nil)
	h = hmac.New(sha256.New, kDate)
	h.Write([]byte(s.region))
	kRegion := h.Sum(nil)
	h = hmac.New(sha256.New, kRegion)
	h.Write([]byte("oss"))
	kService := h.Sum(nil)
	h = hmac.New(sha256.New, kService)
	h.Write([]byte("aliyun_v4_request"))
	kSigning := h.Sum(nil)

	h = hmac.New(sha256.New, kSigning)
	h.Write([]byte(stringToSign))
	signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

	authorization := fmt.Sprintf("%s Credential=%s/%s, Signature=%s", signAlgorithmHeader, s.accessKeyId, scope, signature)

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Date", headerDateStr)

	return nil
}

func escapeQuery(str string) string {
	res := url.QueryEscape(str)
	res = strings.ReplaceAll(res, "+", "%20")
	return res
}

func escapePath(path string) string {
	var buf bytes.Buffer
	for i := 0; i < len(path); i++ {
		c := path[i]
		noEscape := (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' ||
			c == '.' ||
			c == '_' ||
			c == '~' ||
			c == '/'
		if noEscape {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}
