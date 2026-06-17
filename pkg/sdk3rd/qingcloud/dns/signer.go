package dns

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type signer struct {
	accessKeyId     string
	secretAccessKey string
}

func (s *signer) Sign(req *http.Request) error {
	// 签名机制：
	// https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/dns/calling_method/signature/

	date := time.Now().UTC().Format(time.RFC1123)

	verb := req.Method

	canonicalizedResource := "/"
	if req.URL != nil {
		canonicalizedResource = req.URL.Path
		if req.URL.RawQuery != "" {
			values, _ := url.ParseQuery(req.URL.RawQuery)
			canonicalizedResource += "?" + values.Encode()
		}
	}

	stringToSign := verb + "\n" +
		date + "\n" +
		canonicalizedResource

	h := hmac.New(sha256.New, []byte(s.secretAccessKey))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	authorization := fmt.Sprintf("QC-HMAC-SHA256 %s:%s", s.accessKeyId, signature)

	req.Header.Set("Date", date)
	req.Header.Set("Authorization", authorization)

	return nil
}
