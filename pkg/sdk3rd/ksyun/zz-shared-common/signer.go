package openapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type signer struct {
	accessKeyId     string
	secretAccessKey string
	service         string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://docs.ksyun.com/documents/40298

	method := strings.ToUpper(req.Method)

	query := url.Values{}
	if req.URL != nil {
		query = req.URL.Query()
	}

	payload := make(map[string]string)
	if method != http.MethodGet && req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(payloadb, &payload); err != nil {
			return err
		}

		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}

	params := make(map[string]string)
	for k := range query {
		params[k] = query.Get(k)
	}
	for k, v := range payload {
		params[k] = v
	}

	nowUtc := time.Now().UTC()
	timestamp := nowUtc.Format("2006-01-02T15:04:05Z")

	params["Accesskey"] = s.accessKeyId
	params["Service"] = s.service
	params["Timestamp"] = timestamp
	params["SignatureVersion"] = "1.0"
	params["SignatureMethod"] = "HMAC-SHA256"
	paramsKeys := make([]string, 0, len(params))
	for k := range params {
		paramsKeys = append(paramsKeys, k)
	}
	sort.Strings(paramsKeys)

	stringToSign := ""
	for i, k := range paramsKeys {
		if i > 0 {
			stringToSign += "&"
		}

		stringToSign += escapeQuery(k) + "=" + escapeQuery(params[k])
	}

	h := hmac.New(sha256.New, []byte(s.secretAccessKey))
	h.Write([]byte(stringToSign))
	signature := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

	if method == http.MethodGet {
		query.Set("Accesskey", params["Accesskey"])
		query.Set("Service", params["Service"])
		query.Set("Timestamp", params["Timestamp"])
		query.Set("SignatureVersion", params["SignatureVersion"])
		query.Set("SignatureMethod", params["SignatureMethod"])

		req.URL.RawQuery = query.Encode() + "&Signature=" + signature
	} else {
		if _, ok := payload["Action"]; ok {
			query.Set("Action", payload["Action"])
			delete(payload, "Action")
		}

		if _, ok := payload["Version"]; ok {
			query.Set("Version", payload["Version"])
			delete(payload, "Version")
		}

		payload["Accesskey"] = params["Accesskey"]
		payload["Service"] = params["Service"]
		payload["Timestamp"] = params["Timestamp"]
		payload["SignatureVersion"] = params["SignatureVersion"]
		payload["SignatureMethod"] = params["SignatureMethod"]
		payload["Signature"] = signature

		jsonb, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req.URL.RawQuery = query.Encode()
		req.Body = io.NopCloser(bytes.NewReader(jsonb))
		req.ContentLength = int64(len(jsonb))
	}

	return nil
}

func escapeQuery(str string) string {
	res := url.QueryEscape(str)
	res = strings.ReplaceAll(res, "+", "%20")
	return res
}
