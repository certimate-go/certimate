package cert

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestEqualCertificateChainsFromPEM(t *testing.T) {
	certificate := func(serial int64) string {
		t.Helper()
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		template := &x509.Certificate{
			SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "test.example.com"},
			NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		}
		der, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
		if err != nil {
			t.Fatal(err)
		}
		return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	}
	leaf, intermediate, alternate, root := certificate(1), certificate(2), certificate(3), certificate(4)
	chain := leaf + intermediate + root
	for _, tt := range []struct {
		name string
		a, b string
		want bool
	}{
		{"identical", chain, chain, true},
		{"whitespace and CRLF", chain, "\n" + strings.ReplaceAll(chain, "\n", "\r\n") + "\n", true},
		{"same leaf with different intermediate", chain, leaf + alternate + root, false},
		{"same leaf with different root", chain, leaf + intermediate + alternate, false},
		{"added certificate", leaf + intermediate, chain, false},
		{"removed certificate", chain, leaf, false},
		{"reordered certificates", chain, leaf + root + intermediate, false},
		{"different leaf", chain, alternate + intermediate + root, false},
		{"invalid PEM", chain, "not a certificate", false},
		{"invalid DER", chain, leaf + string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("invalid")})), false},
		{"empty", "", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualCertificateChainsFromPEM(tt.a, tt.b); got != tt.want {
				t.Fatalf("EqualCertificateChainsFromPEM() = %v, want %v", got, tt.want)
			}
		})
	}
	if !EqualCertificatesFromPEM(chain, leaf+alternate+root) {
		t.Fatal("leaf-only comparison semantics must remain unchanged for existing callers")
	}
}
