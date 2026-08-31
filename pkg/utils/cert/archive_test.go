package cert

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func makeTestCert(t *testing.T) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com", "www.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))
	return certPEM, keyPEM
}

func TestBuildCertificateArchivePEM(t *testing.T) {
	certPEM, keyPEM := makeTestCert(t)

	data, err := BuildCertificateArchive(certPEM, keyPEM, "example.com", CertificateArchiveOptions{FileFormat: "PEM"})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal("zip invalid:", err)
	}

	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, want := range []string{"example.com.key", "example.com.crt", "example.com (server).pem", "example.com (intermedia).pem"} {
		if !names[want] {
			t.Errorf("zip missing file: %s (got %v)", want, names)
		}
	}
	t.Logf("PEM zip files: %v", names)
}

func TestBuildCertificateArchivePFX(t *testing.T) {
	certPEM, keyPEM := makeTestCert(t)

	data, err := BuildCertificateArchive(certPEM, keyPEM, "example.com", CertificateArchiveOptions{FileFormat: "PFX", PfxPassword: "secret123"})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal("zip invalid:", err)
	}

	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["example.com.pfx"] || !names["README.txt"] {
		t.Errorf("zip missing files: %v", names)
	}
	t.Logf("PFX zip files: %v", names)
}
