package certacme_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	ycdns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	iam "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type b3YandexDNS struct {
	ycdns.UnimplementedDnsZoneServiceServer
	f *b3DNSFixture
}

type b3YandexEndpoints struct {
	endpoint.UnimplementedApiEndpointServiceServer
}

type b3YandexIAM struct {
	iam.UnimplementedIamTokenServiceServer
	f   *b3DNSFixture
	key *rsa.PublicKey
}

func (*b3YandexEndpoints) List(context.Context, *endpoint.ListApiEndpointsRequest) (*endpoint.ListApiEndpointsResponse, error) {
	return &endpoint.ListApiEndpointsResponse{Endpoints: []*endpoint.ApiEndpoint{
		{Id: "iam", Address: "iam.api.cloud.yandex.net:443"},
		{Id: "dns", Address: "dns.api.cloud.yandex.net:443"},
	}}, nil
}

func (s *b3YandexIAM) Create(_ context.Context, request *iam.CreateIamTokenRequest) (*iam.CreateIamTokenResponse, error) {
	_, err := jwt.Parse(request.GetJwt(), func(token *jwt.Token) (any, error) { return s.key, nil }, jwt.WithValidMethods([]string{"PS256"}), jwt.WithIssuer("fixture-service-account"))
	if err != nil {
		s.f.t.Errorf("Yandex service account signature: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid service account signature")
	}
	return &iam.CreateIamTokenResponse{IamToken: "fixture-yandex-dns-token", ExpiresAt: timestamppb.New(time.Now().Add(time.Hour))}, nil
}

func (s *b3YandexDNS) authorized(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	if strings.Join(md.Get("authorization"), "") != "Bearer fixture-yandex-dns-token" {
		s.f.t.Error("Yandex DNS authorization header mismatch")
		return status.Error(codes.Unauthenticated, "invalid token")
	}
	if s.f.mode == "auth-failure" {
		s.f.denied++
		return status.Error(codes.PermissionDenied, "fixture DNS permission denied")
	}
	return nil
}

func (s *b3YandexDNS) List(ctx context.Context, request *ycdns.ListDnsZonesRequest) (*ycdns.ListDnsZonesResponse, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()
	if err := s.authorized(ctx); err != nil {
		return nil, err
	}
	if request.FolderId != "fixture-folder" {
		s.f.t.Errorf("Yandex folder %q", request.FolderId)
		return nil, status.Error(codes.InvalidArgument, "wrong folder")
	}
	return &ycdns.ListDnsZonesResponse{DnsZones: []*ycdns.DnsZone{{Id: "fixture-zone", FolderId: "fixture-folder", Zone: "example.test."}}}, nil
}

func (s *b3YandexDNS) GetRecordSet(ctx context.Context, request *ycdns.GetDnsZoneRecordSetRequest) (*ycdns.RecordSet, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()
	if err := s.authorized(ctx); err != nil {
		return nil, err
	}
	if request.DnsZoneId != "fixture-zone" || request.Type != "TXT" || strings.TrimSuffix(request.Name, ".") != strings.TrimSuffix("_acme-challenge."+s.f.domain, ".example.test") {
		s.f.t.Errorf("unexpected Yandex TXT lookup: %v", request)
		return nil, status.Error(codes.InvalidArgument, "wrong record")
	}
	if s.f.value == "" {
		return nil, status.Error(codes.NotFound, "RecordSet not found")
	}
	return &ycdns.RecordSet{Name: request.Name, Type: "TXT", Ttl: 60, Data: []string{s.f.value}}, nil
}

func (s *b3YandexDNS) UpdateRecordSets(ctx context.Context, request *ycdns.UpdateRecordSetsRequest) (*operation.Operation, error) {
	s.f.mu.Lock()
	defer s.f.mu.Unlock()
	if err := s.authorized(ctx); err != nil {
		return nil, err
	}
	if request.DnsZoneId != "fixture-zone" {
		return nil, status.Error(codes.InvalidArgument, "wrong zone")
	}
	for _, record := range request.Deletions {
		if record.Type != "TXT" || len(record.Data) != 1 || record.Data[0] != s.f.value || s.f.value == "" {
			s.f.t.Error("Yandex deleted unexpected TXT value")
			return nil, status.Error(codes.InvalidArgument, "wrong deletion")
		}
		s.f.value = ""
		s.f.deletes++
	}
	for _, record := range request.Additions {
		if record.Type != "TXT" || strings.TrimSuffix(record.Name, ".") != strings.TrimSuffix("_acme-challenge."+s.f.domain, ".example.test") || len(record.Data) != 1 || record.Data[0] != s.f.expected {
			s.f.t.Error("Yandex uploaded unexpected TXT value")
			return nil, status.Error(codes.InvalidArgument, "wrong addition")
		}
		s.f.value = record.Data[0]
		s.f.creates++
	}
	meta, err := anypb.New(&ycdns.UpdateRecordSetsMetadata{})
	if err != nil {
		return nil, err
	}
	response, err := anypb.New(&ycdns.RecordSetDiff{})
	if err != nil {
		return nil, err
	}
	return &operation.Operation{Id: "fixture-dns-operation", Done: true, Metadata: meta, Result: &operation.Operation_Response{Response: response}}, nil
}

func b3YandexDNSConfig(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
	t.Helper()
	accountKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	account, err := json.Marshal(map[string]string{
		"id": "fixture-key", "service_account_id": "fixture-service-account",
		"private_key": string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(accountKey)})),
	})
	if err != nil {
		t.Fatal(err)
	}
	hosts := []string{"api.cloud.yandex.net", "iam.api.cloud.yandex.net", "dns.api.cloud.yandex.net"}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(103), Subject: pkix.Name{CommonName: "B3 Yandex DNS fixture"}, DNSNames: hosts, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	pair, err := tls.X509KeyPair(certPEM, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	ycdns.RegisterDnsZoneServiceServer(server, &b3YandexDNS{f: f})
	endpoint.RegisterApiEndpointServiceServer(server, &b3YandexEndpoints{})
	iam.RegisterIamTokenServiceServer(server, &b3YandexIAM{f: f, key: &accountKey.PublicKey})
	cloud := httptest.NewUnstartedServer(server)
	cloud.EnableHTTP2 = true
	cloud.TLS = &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12}
	cloud.StartTLS()
	t.Cleanup(cloud.Close)
	t.Cleanup(server.Stop)
	if f.proxyTargets == nil {
		f.proxyTargets = map[string]string{}
	}
	for _, host := range hosts {
		f.proxyTargets[host+":443"] = cloud.Listener.Addr().String()
	}
	f.proxyHosts = append(f.proxyHosts, hosts...)
	f.extraTrust = append(f.extraTrust, certPEM...)
	return b3DNSProviderConfig{Provider: "yandexcloud", Access: map[string]any{"folderId": "fixture-folder", "serviceAccountKey": string(account)}}
}
