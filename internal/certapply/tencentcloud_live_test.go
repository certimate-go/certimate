//go:build tester

package certapply

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestTencentCloudLiveIssuance(t *testing.T) {
	if os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_TEST") != "1" {
		t.Skip("real cloud test disabled; enable CERTIMATE_TENCENTCLOUD_LIVE_TEST=1 only for an authorized domain (consumes quota and changes DNS)")
	}
	for _, name := range []string{
		"CERTIMATE_TENCENTCLOUD_LIVE_SECRET_ID",
		"CERTIMATE_TENCENTCLOUD_LIVE_SECRET_KEY",
		"CERTIMATE_TENCENTCLOUD_LIVE_DOMAIN",
		"CERTIMATE_TENCENTCLOUD_LIVE_EMAIL",
	} {
		if os.Getenv(name) == "" {
			t.Fatalf("required environment variable %s is missing", name)
		}
	}
	projectId := int64(0)
	if value := os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_PROJECT_ID"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 0 {
			t.Fatal("CERTIMATE_TENCENTCLOUD_LIVE_PROJECT_ID must be a nonnegative integer")
		}
		projectId = parsed
	}
	algorithm := os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_KEY_ALGORITHM")
	if algorithm == "" {
		algorithm = "RSA2048"
	}
	request := TencentCloudRequest{
		Domains:       []string{os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_DOMAIN")},
		ContactEmail:  os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_EMAIL"),
		KeySource:     "auto",
		KeyAlgorithm:  algorithm,
		CertificateId: os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_CERTIFICATE_ID"),
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("system CA trust store is required before issuing a real certificate: %v", err)
	}
	client, err := NewTencentCloudClient(TencentCloudConfig{
		SecretId:  os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_SECRET_ID"),
		SecretKey: os.Getenv("CERTIMATE_TENCENTCLOUD_LIVE_SECRET_KEY"),
		ProjectId: projectId,
	}, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	if request.CertificateId == "" {
		t.Logf("requesting a real free certificate for %s using DNS_AUTO and %s", request.Domains[0], algorithm)
	} else {
		t.Logf("resuming real certificate order %s", request.CertificateId)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()
	result, err := client.Obtain(ctx, request, func(id string) error {
		t.Logf("order created: %s; preserve this ID as CERTIMATE_TENCENTCLOUD_LIVE_CERTIFICATE_ID before retrying", id)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CertificateId == "" || (request.CertificateId != "" && result.CertificateId != request.CertificateId) {
		t.Fatal("unexpected certificate order ID")
	}
	pair, err := tls.X509KeyPair([]byte(result.Certificate), []byte(result.PrivateKey))
	if err != nil {
		t.Fatalf("downloaded certificate and private key do not match: %v", err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	intermediates := x509.NewCertPool()
	for _, der := range pair.Certificate[1:] {
		issuer, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		intermediates.AddCert(issuer)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: roots, Intermediates: intermediates, DNSName: request.Domains[0]}); err != nil {
		t.Fatalf("downloaded certificate is not trusted for the requested domain: %v", err)
	}
	t.Logf("verified real certificate order %s: domain=%s algorithm=%s issuer=%s expires=%s chainLength=%d", result.CertificateId, request.Domains[0], algorithm, leaf.Issuer.CommonName, leaf.NotAfter.Format(time.RFC3339), len(pair.Certificate))
}
