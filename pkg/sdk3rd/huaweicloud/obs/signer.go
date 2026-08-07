package obs

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type signer struct {
	accessKeyId     string
	secretAccessKey string
	bucket          string
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://support.huaweicloud.com/api-obs/obs_04_0010.html

	canonicalizedHeaders := ""
	if len(req.Header) > 0 {
		keys := make([]string, 0, len(req.Header))
		for key := range req.Header {
			key = strings.ToLower(key)
			if strings.HasPrefix(key, "x-obs-") {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := strings.TrimSpace(req.Header.Get(key))
			canonicalizedHeaders += key + ":" + escapeQuery(value)
			canonicalizedHeaders += "\n"
		}
	}

	bucketName := s.bucket
	objectName := strings.Trim(req.URL.Path, "/")
	canonicalizedResources := escapePath(fmt.Sprintf("/%s/%s", bucketName, objectName))
	if bucketName == "" && objectName == "" {
		canonicalizedResources = "/"
	}
	if len(req.URL.Query()) > 0 {
		query := req.URL.Query()

		keys := xmaps.Keys(query)
		sort.Strings(keys)

		for i, key := range keys {
			if i == 0 {
				canonicalizedResources += "?"
			} else {
				canonicalizedResources += "&"
			}

			value := query.Get(key)
			if value == "" {
				canonicalizedResources += escapeQuery(key)
			} else {
				canonicalizedResources += escapeQuery(key) + "=" + escapeQuery(value)
			}
		}
	}

	method := strings.ToUpper(req.Method)

	dateStr := time.Now().UTC().Format(http.TimeFormat)

	contentMd5 := req.Header.Get("Content-MD5")
	contentType := req.Header.Get("Content-Type")

	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s%s", method, contentMd5, contentType, dateStr, canonicalizedHeaders, canonicalizedResources)

	h := hmac.New(sha1.New, []byte(s.secretAccessKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	authorization := fmt.Sprintf("OBS %s:%s", s.accessKeyId, signature)

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Date", dateStr)

	return nil
}

func escapeQuery(str string) string {
	res := url.QueryEscape(str)
	res = strings.ReplaceAll(res, "%7E", "~")
	res = strings.ReplaceAll(res, "%2F", "/")
	res = strings.ReplaceAll(res, "%20", "+")
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
