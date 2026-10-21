package tencentcloudssl

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"
)

type uploadTestTransport func(*http.Request) (*http.Response, error)

func (f uploadTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestUploadPreservesReplacementChain(t *testing.T) {
	certificate := func(serial int64) string {
		t.Helper()
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		template := &x509.Certificate{SerialNumber: big.NewInt(serial), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
		der, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
		if err != nil {
			t.Fatal(err)
		}
		return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	}
	leaf, oldIssuer, newIssuer := certificate(1), certificate(2), certificate(3)
	wantedChain := leaf + newIssuer
	for _, tt := range []struct {
		name        string
		duplicate   bool
		existing    string
		readDenied  bool
		retryFails  bool
		retryReuses bool
		wantCalls   int
		wantID      string
		wantError   bool
	}{
		{name: "new upload", wantCalls: 1, wantID: "new-id"},
		{name: "same full chain reuses existing", duplicate: true, existing: wantedChain, wantCalls: 2, wantID: "old-id"},
		{name: "same chain with CRLF", duplicate: true, existing: strings.ReplaceAll(wantedChain, "\n", "\r\n"), wantCalls: 2, wantID: "old-id"},
		{name: "same leaf with changed issuer", duplicate: true, existing: leaf + oldIssuer, wantCalls: 3, wantID: "new-id"},
		{name: "missing chain", duplicate: true, wantCalls: 3, wantID: "new-id"},
		{name: "no additional read permission", duplicate: true, readDenied: true, wantCalls: 3, wantID: "new-id"},
		{name: "replacement upload failure", duplicate: true, retryFails: true, wantCalls: 3, wantError: true},
		{name: "cloud still returns old certificate", duplicate: true, retryReuses: true, wantCalls: 3, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewCertmgr(&CertmgrConfig{SecretId: "test-id", SecretKey: "test-secret"})
			if err != nil {
				t.Fatal(err)
			}
			manager.SetLogger(nil)
			calls := 0
			manager.sdkClient.WithHttpTransport(uploadTestTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				var response map[string]any
				switch calls {
				case 1, 3:
					if req.Header.Get("X-TC-Action") != "UploadCertificate" || payload["Repeatable"] != (calls == 3) {
						t.Fatalf("unexpected upload action or deduplication flag at call %d", calls)
					}
					if payload["CertificatePublicKey"] != wantedChain || payload["CertificatePrivateKey"] != "matching-test-key" {
						t.Fatal("retry must preserve the requested full chain and key")
					}
					response = map[string]any{"CertificateId": "new-id"}
					if (calls == 1 && tt.duplicate) || (calls == 3 && tt.retryReuses) {
						response = map[string]any{"CertificateId": "old-id", "RepeatCertId": "old-id"}
					}
					if calls == 3 && tt.retryFails {
						response = map[string]any{"Error": map[string]string{"Code": "AuthFailure.UnauthorizedOperation", "Message": "upload denied"}}
					}
				case 2:
					if req.Header.Get("X-TC-Action") != "DescribeCertificateDetail" || payload["CertificateId"] != "old-id" {
						t.Fatal("must inspect the duplicate certificate before reusing it")
					}
					response = map[string]any{"CertificatePublicKey": tt.existing}
					if tt.readDenied {
						response = map[string]any{"Error": map[string]string{"Code": "AuthFailure.UnauthorizedOperation", "Message": "read denied"}}
					}
				default:
					t.Fatalf("unexpected API call %d", calls)
				}
				body, err := json.Marshal(map[string]any{"Response": response})
				if err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
			}))
			result, err := manager.Upload(context.Background(), wantedChain, "matching-test-key")
			if (err != nil) != tt.wantError {
				t.Fatalf("Upload() error = %v, want error = %v", err, tt.wantError)
			}
			if !tt.wantError && result.CertId != tt.wantID {
				t.Fatalf("Upload() ID = %q, want %q", result.CertId, tt.wantID)
			}
			if calls != tt.wantCalls {
				t.Fatalf("API calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}
