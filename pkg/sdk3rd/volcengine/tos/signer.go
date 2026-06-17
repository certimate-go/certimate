package tos

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
	secretAccessKey string
	region          string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://www.volcengine.com/docs/6349/74839
	// https://docs.byteplus.com/en/docs/tos/reference-signature-mechanism_1

	method := strings.ToUpper(req.Method)

	nowUtc := time.Now().UTC()
	headerDateStr := nowUtc.Format(http.TimeFormat)
	requestDateStr := nowUtc.Format("20060102T150405Z")
	credentialDateStr := nowUtc.Format("20060102")

	canonicalUrl := escapePath(req.URL.Path)
	if canonicalUrl == "" {
		canonicalUrl = "/"
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
			canonicalQueryStr += escapeQuery(key) + "=" + escapeQuery(value)
		}
	}

	canonicalHeaders := ""
	signedHeaders := ""
	if len(req.Header) > 0 || req.Host != "" {
		if req.Header.Get("Host") == "" {
			req.Header.Set("Host", req.Host)
		}
		if req.Header.Get("X-TOS-Date") == "" {
			req.Header.Set("X-TOS-Date", requestDateStr)
		}

		keys := make([]string, 0, len(req.Header))
		for key := range req.Header {
			key = strings.ToLower(key)
			if strings.HasPrefix(key, "x-tos-") || key == "host" {
				keys = append(keys, key)
			}
			if key == "content-type" && req.Header.Get("X-TOS-Content-SHA256") != "" {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)

		for i, key := range keys {
			if i > 0 {
				canonicalHeaders += "\n"
				signedHeaders += ";"
			}

			value := strings.TrimSpace(req.Header.Get(key))
			canonicalHeaders += key + ":" + value
			signedHeaders += key
		}

		canonicalHeaders += "\n"
	}

	hashedPayload := req.Header.Get("X-TOS-Content-SHA256")
	if hashedPayload == "" {
		temp := sha256.Sum256([]byte{})
		hashedPayload = strings.ToLower(hex.EncodeToString(temp[:]))
	}

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, canonicalUrl, canonicalQueryStr, canonicalHeaders, signedHeaders, hashedPayload)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHashHex := strings.ToLower(hex.EncodeToString(canonicalRequestHash[:]))

	const signAlgorithmHeader = "TOS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/tos/request", credentialDateStr, s.region)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", signAlgorithmHeader, requestDateStr, credentialScope, canonicalRequestHashHex)

	var h hash.Hash
	h = hmac.New(sha256.New, []byte(s.secretAccessKey))
	h.Write([]byte(credentialDateStr))
	kDate := h.Sum(nil)
	h = hmac.New(sha256.New, kDate)
	h.Write([]byte(s.region))
	kRegion := h.Sum(nil)
	h = hmac.New(sha256.New, kRegion)
	h.Write([]byte("tos"))
	kService := h.Sum(nil)
	h = hmac.New(sha256.New, kService)
	h.Write([]byte("request"))
	kSigning := h.Sum(nil)

	h = hmac.New(sha256.New, kSigning)
	h.Write([]byte(stringToSign))
	signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s", signAlgorithmHeader, s.accessKeyId, credentialScope, signedHeaders, signature)

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
