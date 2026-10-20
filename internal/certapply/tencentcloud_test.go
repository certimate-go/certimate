package certapply

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestTencentCloudClient(t *testing.T, handler func(string, map[string]any) map[string]any) (*TencentCloudClient, *bytes.Buffer) {
	t.Helper()
	logs := new(bytes.Buffer)
	client, err := NewTencentCloudClient(TencentCloudConfig{SecretId: "test-id", SecretKey: "test-secret", ProjectId: 42}, slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	if err != nil {
		t.Fatal(err)
	}
	client.pollInterval = time.Millisecond
	client.client.(*ssl.Client).WithHttpTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "ssl.tencentcloudapi.com" || req.Header.Get("Authorization") == "" {
			t.Errorf("expected a signed Tencent SSL request, got %s", req.URL)
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		response := handler(req.Header.Get("X-TC-Action"), body)
		data, err := json.Marshal(map[string]any{"Response": response})
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(data))}, nil
	}))
	return client, logs
}

func testCertificatePEM(t *testing.T, algorithm, domain string, expires time.Time) (string, string) {
	t.Helper()
	var key crypto.Signer
	var err error
	if algorithm == "RSA2048" {
		key, err = rsa.GenerateKey(rand.Reader, 2048)
	} else {
		key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: domain}, DNSNames: []string{domain},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: expires,
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "Test CA"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(365 * 24 * time.Hour)}
	issuerDER, err := x509.CreateCertificate(rand.Reader, issuer, issuer, issuerKey.Public(), issuerKey)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuer, key.Public(), issuerKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	bundle := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: issuerDER})...)
	return string(bundle), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
}

func testCertificateArchive(t *testing.T, files map[string]string) string {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(file, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func validTencentCloudRequest() TencentCloudRequest {
	return TencentCloudRequest{Domains: []string{"test.example.com"}, KeySource: "auto", KeyAlgorithm: "RSA2048", ContactEmail: "test@example.com"}
}

func TestTencentCloudObtain(t *testing.T) {
	for _, algorithm := range []string{"RSA2048", "EC256"} {
		t.Run(algorithm, func(t *testing.T) {
			cert, key := testCertificatePEM(t, algorithm, "test.example.com", time.Now().Add(90*24*time.Hour))
			var actions []string
			statuses := []int{4, 0, 0, 1}
			persisted := false
			client, logs := newTestTencentCloudClient(t, func(action string, body map[string]any) map[string]any {
				actions = append(actions, action)
				switch action {
				case "ApplyCertificate":
					for field, expected := range map[string]any{"DomainName": "test.example.com", "DvAuthMethod": "DNS_AUTO", "PackageType": "83", "ValidityPeriod": "3", "ProjectId": float64(42), "ContactEmail": "test@example.com", "DeleteDnsAutoRecord": true} {
						if body[field] != expected {
							t.Errorf("%s = %v, want %v", field, body[field], expected)
						}
					}
					algo, parameter := "RSA", "2048"
					if algorithm == "EC256" {
						algo, parameter = "ECC", "prime256v1"
					}
					if body["CsrEncryptAlgo"] != algo || body["CsrKeyParameter"] != parameter {
						t.Errorf("incorrect key parameters: %v", body)
					}
					if body["PackageId"] != nil || body["DnsNames"] != nil || body["OldCertificateId"] != nil {
						t.Errorf("unexpected paid/multidomain/renewal parameters: %v", body)
					}
					return map[string]any{"CertificateId": "order-1"}
				case "DescribeCertificateDetail":
					if !persisted {
						t.Fatal("order must be persisted before waiting")
					}
					if body["CertificateId"] != "order-1" {
						t.Fatal("wrong certificate ID")
					}
					status := statuses[0]
					statuses = statuses[1:]
					return map[string]any{"Status": status}
				case "DownloadCertificate":
					if body["CertificateId"] != "order-1" {
						t.Fatal("wrong download certificate ID")
					}
					return map[string]any{"ContentType": "application/zip", "Content": testCertificateArchive(t, map[string]string{"Nginx/test.example.com_bundle.crt": cert, "Nginx/test.example.com.key": key})}
				case "CompleteCertificate":
					if body["CertificateId"] != "order-1" {
						t.Fatal("wrong certificate ID")
					}
					return map[string]any{"CertificateId": "order-1"}
				default:
					t.Fatalf("unexpected action: %s", action)
					return nil
				}
			})
			request := validTencentCloudRequest()
			request.KeyAlgorithm = algorithm
			result, err := client.Obtain(context.Background(), request, func(id string) error {
				persisted = id == "order-1"
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Certificate != cert || result.PrivateKey != key || result.CertificateId != "order-1" {
				t.Fatal("incorrect certificate result")
			}
			if !reflect.DeepEqual(actions, []string{"ApplyCertificate", "DescribeCertificateDetail", "DescribeCertificateDetail", "CompleteCertificate", "DescribeCertificateDetail", "DescribeCertificateDetail", "DownloadCertificate"}) {
				t.Fatalf("unexpected action sequence: %v", actions)
			}
			if strings.Contains(logs.String(), "PRIVATE KEY") || strings.Contains(logs.String(), "test-secret") || strings.Contains(logs.String(), "BEGIN CERTIFICATE") {
				t.Fatal("sensitive PEM/credentials were logged")
			}
		})
	}
}

func TestTencentCloudTransitionalStates(t *testing.T) {
	for _, status := range []uint64{4, 6, 9, 11, 15} {
		ctx, cancel := context.WithCancel(context.Background())
		client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
			if action != "DescribeCertificateDetail" {
				t.Fatalf("unexpected action for status %d: %s", status, action)
			}
			cancel()
			return map[string]any{"Status": status}
		})
		request := validTencentCloudRequest()
		request.CertificateId = "transitioning-order"
		_, err := client.Obtain(ctx, request, nil)
		cancel()
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("status %d must remain pending, got %v", status, err)
		}
	}
}

func TestTencentCloudResumeAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client, _ := newTestTencentCloudClient(t, func(action string, body map[string]any) map[string]any {
		if action != "DescribeCertificateDetail" || body["CertificateId"] != "pending-order" {
			t.Fatalf("resuming must not apply again: %s", action)
		}
		cancel()
		return map[string]any{"Status": 4}
	})
	request := validTencentCloudRequest()
	request.CertificateId = "pending-order"
	_, err := client.Obtain(ctx, request, nil)
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "pending-order") {
		t.Fatalf("expected cancellation with order ID, got %v", err)
	}
}

func TestTencentCloudDeadline(t *testing.T) {
	client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
		if action == "ApplyCertificate" {
			return map[string]any{"CertificateId": "slow-order"}
		}
		return map[string]any{"Status": 4}
	})
	client.pollInterval = time.Hour
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := client.Obtain(ctx, validTencentCloudRequest(), nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}

func TestTencentCloudAPIAndResponseErrors(t *testing.T) {
	for _, test := range []struct {
		name, action string
		response     map[string]any
		want         string
	}{
		{"permission", "ApplyCertificate", map[string]any{"Error": map[string]string{"Code": "AuthFailure.UnauthorizedOperation", "Message": "denied"}}, "AuthFailure.UnauthorizedOperation"},
		{"quota", "ApplyCertificate", map[string]any{"Error": map[string]string{"Code": "LimitExceeded.ExceedsFreeLimit", "Message": "quota exhausted"}}, "LimitExceeded.ExceedsFreeLimit"},
		{"empty order", "ApplyCertificate", map[string]any{}, "empty certificate ID"},
		{"missing status", "DescribeCertificateDetail", map[string]any{}, "no status"},
		{"rejected", "DescribeCertificateDetail", map[string]any{"Status": 2, "StatusMsg": "CA rejected"}, "CA rejected"},
		{"revoked", "DescribeCertificateDetail", map[string]any{"Status": 10}, "status 10"},
		{"unknown state", "DescribeCertificateDetail", map[string]any{"Status": 999}, "status 999"},
		{"missing archive", "DownloadCertificate", map[string]any{}, "no archive"},
		{"download permission", "DownloadCertificate", map[string]any{"Error": map[string]string{"Code": "AuthFailure.UnauthorizedOperation", "Message": "denied"}}, "ssl:DownloadCertificate"},
		{"validation permission", "CompleteCertificate", map[string]any{"Error": map[string]string{"Code": "AuthFailure.UnauthorizedOperation", "Message": "denied"}}, "ssl:CompleteCertificate"},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
				if action == test.action {
					return test.response
				}
				if action == "ApplyCertificate" {
					return map[string]any{"CertificateId": "order-error"}
				}
				if test.action == "DownloadCertificate" {
					return map[string]any{"Status": 1}
				}
				return map[string]any{"Status": 0}
			})
			_, err := client.Obtain(context.Background(), validTencentCloudRequest(), nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("want %s, got %v", test.want, err)
			}
			if test.name == "rejected" {
				var statusErr *TencentCloudStatusError
				if !errors.As(err, &statusErr) || statusErr.CertificateId != "order-error" {
					t.Fatal("missing typed terminal status")
				}
			}
		})
	}
}

func TestTencentCloudPersistFailure(t *testing.T) {
	client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
		if action != "ApplyCertificate" {
			t.Fatal("must not continue when order persistence fails")
		}
		return map[string]any{"CertificateId": "recoverable-id"}
	})
	_, err := client.Obtain(context.Background(), validTencentCloudRequest(), func(string) error { return errors.New("database unavailable") })
	if err == nil || !strings.Contains(err.Error(), "recoverable-id") {
		t.Fatalf("expected recoverable order ID, got %v", err)
	}
}

func TestTencentCloudRejectInvalidRequest(t *testing.T) {
	for _, change := range []func(*TencentCloudRequest){
		func(r *TencentCloudRequest) { r.Domains = nil },
		func(r *TencentCloudRequest) { r.Domains = []string{"one.example.com", "two.example.com"} },
		func(r *TencentCloudRequest) { r.Domains = []string{"*.example.com"} },
		func(r *TencentCloudRequest) { r.Domains = []string{"127.0.0.1"} },
		func(r *TencentCloudRequest) { r.Domains = []string{"https://example.com"} },
		func(r *TencentCloudRequest) { r.Domains = []string{"a..example.com"} },
		func(r *TencentCloudRequest) { r.Domains = []string{"-bad.example.com"} },
		func(r *TencentCloudRequest) { r.IPAddrs = []string{"::1"} },
		func(r *TencentCloudRequest) { r.KeyAlgorithm = "RSA4096" },
		func(r *TencentCloudRequest) { r.KeyAlgorithm = "EC384" },
		func(r *TencentCloudRequest) { r.KeySource = "custom" },
		func(r *TencentCloudRequest) { r.KeySource = "reuse" },
		func(r *TencentCloudRequest) { r.ContactEmail = "not an email" },
	} {
		client, _ := newTestTencentCloudClient(t, func(string, map[string]any) map[string]any { t.Fatal("invalid request reached API"); return nil })
		request := validTencentCloudRequest()
		change(&request)
		if _, err := client.Obtain(context.Background(), request, nil); err == nil {
			t.Fatalf("accepted invalid request: %+v", request)
		}
	}
	for _, config := range []TencentCloudConfig{{}, {SecretId: "id"}, {SecretId: "id", SecretKey: "key", ProjectId: -1}} {
		if _, err := NewTencentCloudClient(config, nil); err == nil {
			t.Fatalf("accepted invalid credentials: %+v", config)
		}
	}
}

func TestTencentCloudArchiveValidation(t *testing.T) {
	cert, key := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(time.Hour))
	leaf, _ := pem.Decode([]byte(cert))
	otherCert, _ := testCertificatePEM(t, "EC256", "other.example.com", time.Now().Add(time.Hour))
	_, otherIssuer := pem.Decode([]byte(otherCert))
	request := validTencentCloudRequest()
	request.KeyAlgorithm = "EC256"
	for name, archive := range map[string]string{
		"invalid base64":  "!",
		"invalid zip":     base64.StdEncoding.EncodeToString([]byte("not a ZIP archive")),
		"missing bundle":  testCertificateArchive(t, map[string]string{"Nginx/test.key": key}),
		"missing key":     testCertificateArchive(t, map[string]string{"Nginx/test_bundle.pem": cert}),
		"leaf only":       testCertificateArchive(t, map[string]string{"Nginx/test_bundle.crt": string(pem.EncodeToMemory(leaf)), "Nginx/test.key": key}),
		"wrong chain":     testCertificateArchive(t, map[string]string{"Nginx/test_bundle.crt": string(pem.EncodeToMemory(leaf)) + string(otherIssuer), "Nginx/test.key": key}),
		"oversized entry": testCertificateArchive(t, map[string]string{"Nginx/test_bundle.crt": strings.Repeat("x", (2<<20)+1), "Nginx/test.key": key}),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := tencentCloudArchiveResult("archive-order", archive, request); err == nil {
				t.Fatal("accepted invalid archive")
			}
		})
	}
}

func TestTencentCloudValidationRace(t *testing.T) {
	cert, key := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(time.Hour))
	polls := 0
	client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
		switch action {
		case "ApplyCertificate":
			return map[string]any{"CertificateId": "racing-order"}
		case "DescribeCertificateDetail":
			polls++
			if polls == 1 {
				return map[string]any{"Status": 13}
			}
			return map[string]any{"Status": 1}
		case "CompleteCertificate":
			return map[string]any{"Error": map[string]string{"Code": "FailedOperation.CertificateNotFoundOrAudited", "Message": "already issued"}}
		case "DownloadCertificate":
			return map[string]any{"Content": testCertificateArchive(t, map[string]string{"Nginx/test_bundle.pem": cert, "Nginx/test.key": key})}
		default:
			t.Fatalf("unexpected action: %s", action)
			return nil
		}
	})
	request := validTencentCloudRequest()
	request.KeyAlgorithm = "EC256"
	if _, err := client.Obtain(context.Background(), request, nil); err != nil {
		t.Fatal(err)
	}
}

func TestTencentCloudCertificateValidation(t *testing.T) {
	cert, key := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(time.Hour))
	otherCert, otherKey := testCertificatePEM(t, "EC256", "other.example.com", time.Now().Add(time.Hour))
	expiredCert, expiredKey := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(-time.Minute))
	for _, test := range []struct{ name, cert, key, algorithm string }{
		{"invalid PEM", "bad certificate", key, "EC256"},
		{"wrong key", cert, otherKey, "EC256"},
		{"wrong domain", otherCert, otherKey, "EC256"},
		{"wrong algorithm", cert, key, "RSA2048"},
		{"expired", expiredCert, expiredKey, "EC256"},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
				if action == "ApplyCertificate" {
					return map[string]any{"CertificateId": "invalid-order"}
				}
				if action == "DownloadCertificate" {
					return map[string]any{"Content": testCertificateArchive(t, map[string]string{"Nginx/test.example.com_bundle.pem": test.cert, "Nginx/test.example.com.key": test.key})}
				}
				return map[string]any{"Status": 1}
			})
			request := validTencentCloudRequest()
			request.KeyAlgorithm = test.algorithm
			if _, err := client.Obtain(context.Background(), request, nil); err == nil {
				t.Fatal("accepted invalid certificate")
			}
		})
	}
}
