package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sigv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type signer struct {
	accessKeyId     string
	secretAccessKey string
	service         string
	region          string
}

func NewSigner(ak, sk, service, region string) *signer {
	return &signer{
		accessKeyId:     ak,
		secretAccessKey: sk,
		service:         service,
		region:          region,
	}
}

func (s *signer) Sign(req *http.Request) error {
	// API 签名机制：
	// https://github.com/aws/smithy-go/blob/a4c9efcda6aa54c75d1a130d1320a2709eebf51d/aws-http-auth/sigv4/sigv4.go

	payload := ([]byte)(nil)
	if req.Body != nil {
		payloadb, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		payload = payloadb
		req.Body = io.NopCloser(bytes.NewReader(payloadb))
	}

	payloadHash := sha256.Sum256(payload)
	payloadHashEncoded := base64.StdEncoding.EncodeToString(payloadHash[:])

	ctx := req.Context()
	cred := aws.Credentials{AccessKeyID: s.accessKeyId, SecretAccessKey: s.secretAccessKey}
	now := time.Now()
	if err := sigv4.NewSigner().SignHTTP(ctx, cred, req, payloadHashEncoded, s.service, s.region, now); err != nil {
		return err
	}

	return nil
}
