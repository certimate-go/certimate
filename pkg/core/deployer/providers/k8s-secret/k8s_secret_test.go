package k8ssecret

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const testKubeConfig = `apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: https://127.0.0.1:6443
      insecure-skip-tls-verify: true
contexts:
  - name: test
    context:
      cluster: test
      user: test
current-context: test
users:
  - name: test
    user:
      token: test-token
`

// Secrets live in the core API group, which is served under "/api".
// If APIPath is left empty, rest.RESTClientFor builds requests against
// "/v1/namespaces/..." and the API server answers 404
// ("the server could not find the requested resource").
func TestCreateK8sClientSetsCoreAPIPath(t *testing.T) {
	client, err := createK8sClient(testKubeConfig)
	if err != nil {
		t.Fatalf("createK8sClient() returned an unexpected error: %v", err)
	}

	const want = "/api/v1"
	if got := client.Get().URL().Path; !strings.HasPrefix(got, want) {
		t.Errorf("request path = %q, want it to start with %q", got, want)
	}
}

// A POST (create) must target the collection path
// "/api/v1/namespaces/<ns>/secrets"; chaining .Name() on the Post builder
// makes the request target the named path ".../secrets/<name>", which the
// API server rejects with 405
// ("the server does not allow this method on the requested resource").
func TestDeployPostsToCollectionPathWhenSecretIsMissing(t *testing.T) {
	// Given: an API server that reports the secret as missing.
	var mu sync.Mutex
	var postPath string
	postCount := 0
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","message":"secrets not found","reason":"NotFound","code":404}`))
		case http.MethodPost:
			mu.Lock()
			postPath, postCount = r.URL.Path, postCount+1
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.Error(w, "unexpected method: "+r.Method, http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	kubeConfig := strings.Replace(testKubeConfig, "https://127.0.0.1:6443", ts.URL, 1)
	deployer, err := NewDeployer(&DeployerConfig{
		KubeConfig:          kubeConfig,
		Namespace:           "test-ns",
		SecretName:          "test-secret",
		SecretType:          "kubernetes.io/tls",
		SecretDataKeyForCrt: "tls.crt",
		SecretDataKeyForKey: "tls.key",
	})
	if err != nil {
		t.Fatalf("NewDeployer() returned an unexpected error: %v", err)
	}
	deployer.SetLogger(slog.New(slog.DiscardHandler))

	certPEM, privkeyPEM := mustTestCertificatePEM(t)

	// When: deploying to the (missing) secret.
	if _, err := deployer.Deploy(context.Background(), certPEM, privkeyPEM); err != nil {
		t.Fatalf("Deploy() returned an unexpected error: %v", err)
	}

	// Then: the POST went to the collection path, without the secret name.
	mu.Lock()
	defer mu.Unlock()
	if postCount != 1 {
		t.Fatalf("POST request count = %d, want 1", postCount)
	}
	const wantPath = "/api/v1/namespaces/test-ns/secrets"
	if postPath != wantPath {
		t.Errorf("POST path = %q, want %q", postPath, wantPath)
	}
}

func mustTestCertificatePEM(t *testing.T) (certPEM, privkeyPEM string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}

	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	privkeyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, privkeyPEM
}
