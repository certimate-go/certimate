package ibmc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeploy_UsesRedfishSessionAndImportsCertificate(t *testing.T) {
	const token = "session-token"
	var imported, closed, restarted int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redfish/v1/SessionService/Sessions":
			if r.Method != http.MethodPost {
				t.Errorf("session method = %s", r.Method)
			}
			w.Header().Set("X-Auth-Token", token)
			w.Header().Set("Location", "/redfish/v1/SessionService/Sessions/1")
			w.WriteHeader(http.StatusCreated)
		case "/redfish/v1/Managers":
			if r.Method != http.MethodGet {
				t.Errorf("managers method = %s", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Members": []map[string]string{{"@odata.id": "/redfish/v1/Managers/1"}}})
		case "/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HttpsCert.ImportServerCertificate":
			t.Errorf("unexpected legacy import endpoint used")
		case "/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HttpsCert.ImportCustomCertificate":
			if r.Header.Get("X-Auth-Token") != token {
				t.Errorf("missing session token")
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode payload: %v", err)
			}
			if payload["Password"] == "" {
				t.Error("PKCS#12 password must not be empty")
			}
			if _, err := base64.StdEncoding.DecodeString(payload["Certificate"]); err != nil {
				t.Errorf("certificate is not base64: %v", err)
			}
			imported++
			w.WriteHeader(http.StatusOK)
		case "/redfish/v1/Managers/1/Actions/Manager.Reset":
			var payload map[string]string
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if payload["ResetType"] != "ForceRestart" {
				t.Errorf("unexpected reset payload: %#v", payload)
			}
			restarted++
			w.WriteHeader(http.StatusOK)
		case "/redfish/v1/SessionService/Sessions/1":
			if r.Method == http.MethodDelete && r.Header.Get("X-Auth-Token") == token {
				closed++
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	certPEM, keyPEM := testCertificate(t)
	deployer, err := NewDeployer(&DeployerConfig{Endpoint: server.URL + "\n" + server.URL, Username: "admin", Password: "secret", RestartAfterImport: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := deployer.Deploy(context.Background(), certPEM, keyPEM); err != nil {
		t.Fatal(err)
	}
	if imported != 2 {
		t.Fatalf("certificate import calls = %d, want 2", imported)
	}
	if restarted != 2 {
		t.Fatalf("iBMC restart calls = %d, want 2", restarted)
	}
	if closed != 2 {
		t.Fatalf("Redfish session close calls = %d, want 2", closed)
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := map[string]string{
		"192.0.2.10":                     "https://192.0.2.10",
		"192.0.2.10:8443/":               "https://192.0.2.10:8443",
		"2001:db8::10":                   "https://[2001:db8::10]",
		"[2001:db8::10]:8443/":           "https://[2001:db8::10]:8443",
		"http://192.0.2.10/":             "http://192.0.2.10",
		"https://[2001:db8::10]/redfish": "https://[2001:db8::10]/redfish",
	}
	for input, want := range tests {
		got, err := normalizeEndpoint(input)
		if err != nil {
			t.Errorf("normalizeEndpoint(%q): %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("normalizeEndpoint(%q) = %q, want %q", input, got, want)
		}
	}
}

func testCertificate(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "example.test"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return string(certPEM), string(keyPEM)
}
