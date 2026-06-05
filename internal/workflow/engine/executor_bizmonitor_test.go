package engine

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExecRetrieveCertificates_SendsSNIFromDomainWhenHostIsIP(t *testing.T) {
	const domain = "monitor.example.com"
	const defaultCN = "default.invalid"

	domainCert := mustGenerateTLSCertificate(t, domain, []string{domain})
	defaultCert := mustGenerateTLSCertificate(t, defaultCN, []string{defaultCN})

	var receivedSNI string
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.TLS = &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			receivedSNI = hello.ServerName
			if hello.ServerName == domain {
				return &tls.Config{Certificates: []tls.Certificate{domainCert}}, nil
			}
			return &tls.Config{Certificates: []tls.Certificate{defaultCert}}, nil
		},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	host, portStr, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	if net.ParseIP(host) == nil {
		t.Fatalf("expected ip host, got %q", host)
	}

	execCtx := (&NodeExecutionContext{}).SetContext(context.Background())
	ne := &bizMonitorNodeExecutor{nodeExecutor: nodeExecutor{}}

	certs, err := ne.execRetrieveCertificates(execCtx, net.JoinHostPort(host, portStr), domain, "/")
	if err != nil {
		t.Fatalf("execRetrieveCertificates: %v", err)
	}
	if len(certs) == 0 {
		t.Fatal("expected peer certificates")
	}

	if receivedSNI != domain {
		t.Fatalf("server received SNI %q, want %q", receivedSNI, domain)
	}
	if certs[0].Subject.CommonName != domain {
		t.Fatalf("peer certificate CN %q, want %q", certs[0].Subject.CommonName, domain)
	}
	if certs[0].VerifyHostname(domain) != nil {
		t.Fatalf("certificate does not match domain %q", domain)
	}
}

func mustGenerateTLSCertificate(t *testing.T, commonName string, dnsNames []string) tls.Certificate {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     dnsNames,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certPEM := tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  privateKey,
	}
	return certPEM
}
