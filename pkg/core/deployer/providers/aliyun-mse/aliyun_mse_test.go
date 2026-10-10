package aliyunmse

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alimse "github.com/alibabacloud-go/mse-20190531/v5/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/certimate-go/certimate/pkg/core"
)

type testCertmgr struct {
	core.Certmgr
	uploads    int
	err        error
	identifier string
}

func (c *testCertmgr) Upload(context.Context, string, string) (*core.CertmgrUploadResult, error) {
	c.uploads++
	return &core.CertmgrUploadResult{ExtendedData: map[string]any{"CertIdWithRegion": c.identifier}}, c.err
}

func newTestDeployer(t *testing.T, config *DeployerConfig, handler http.HandlerFunc) (*Deployer, *testCertmgr) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := alimse.NewClient(&aliopen.Config{
		AccessKeyId:     tea.String("test-key"),
		AccessKeySecret: tea.String("test-secret"),
		Endpoint:        tea.String(strings.TrimPrefix(server.URL, "http://")),
		Protocol:        tea.String("HTTP"),
	})
	if err != nil {
		t.Fatal(err)
	}
	certmgr := &testCertmgr{identifier: "123-cn-hangzhou"}
	return &Deployer{config: config, logger: slog.New(slog.DiscardHandler), sdkClient: client, sdkCertmgr: certmgr}, certmgr
}

func testCertificate(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{"example.com", "*.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func writeResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"Success": true, "Code": 200, "Data": data})
}

func TestNewDeployerValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config *DeployerConfig
		want   string
	}{
		{"nil", nil, "nil"},
		{"region", &DeployerConfig{}, "region"},
		{"gateway", &DeployerConfig{Region: "cn-hangzhou"}, "gatewayId"},
		{"domain", &DeployerConfig{Region: "cn-hangzhou", GatewayId: "gw-test"}, "domain"},
		{"wildcard domain", &DeployerConfig{Region: "cn-hangzhou", GatewayId: "gw-test", DomainMatchPattern: "wildcard"}, "domain"},
		{"pattern", &DeployerConfig{Region: "cn-hangzhou", GatewayId: "gw-test", DomainMatchPattern: "invalid"}, "unsupported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewDeployer(tc.config); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
	_, err := NewDeployer(&DeployerConfig{
		AccessKeyId: "test-key", AccessKeySecret: "test-secret",
		Region: "cn-hangzhou", GatewayId: "gw-test", DomainMatchPattern: "certsan",
	})
	if err != nil {
		t.Fatalf("certificate matching should not require a domain: %v", err)
	}
}

func TestDeployDomainMatching(t *testing.T) {
	certPEM := testCertificate(t)
	for _, tc := range []struct {
		name, pattern, domain, wantError string
		forceHttps                       bool
		wantIDs                          []int
	}{
		{name: "default exact", domain: "api.example.com", wantIDs: []int{1}},
		{name: "exact case insensitive", pattern: "exact", domain: " API.Example.COM ", wantIDs: []int{1}},
		{name: "exact wildcard entry", pattern: "exact", domain: "*.example.com", wantIDs: []int{3}},
		{name: "wildcard", pattern: "wildcard", domain: "*.example.com", wantIDs: []int{1, 2, 3, 6}},
		{name: "certificate SAN", pattern: "certsan", wantIDs: []int{1, 2, 3, 4, 6}},
		{name: "force HTTPS", pattern: "certsan", forceHttps: true, wantIDs: []int{1, 2, 3, 4, 6}},
		{name: "not covered", pattern: "exact", domain: "nested.api.example.com", wantError: "does not cover"},
		{name: "unknown domain", pattern: "exact", domain: "missing.example.com", wantError: "matching domains"},
		{name: "HTTP domain", pattern: "exact", domain: "http.example.com", wantIDs: []int{6}},
		{name: "fallback domain", pattern: "exact", domain: "*", wantError: "matching domains"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var updated []int
			lists := 0
			settings := map[string]string{
				"Http2": "globalConfig", "TlsMin": "TLS 1.2", "TlsMax": "TLS 1.3",
				"MtlsEnabled": "true", "CaCertIdentifier": "456-cn-hangzhou",
				"TlsCipherSuitesConfigJSON": `{"ConfigType":"CUSTOM","TlsCipherSuites":["ECDHE-RSA-AES128-GCM-SHA256"]}`,
			}
			d, certmgr := newTestDeployer(t, &DeployerConfig{
				GatewayId: "gw-test", DomainMatchPattern: tc.pattern, Domain: tc.domain, ForceHttps: tc.forceHttps,
			}, func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				if r.Form.Get("GatewayUniqueId") != "gw-test" {
					t.Error("missing GatewayUniqueId")
				}
				switch r.Header.Get("x-acs-action") {
				case "ListGatewayDomain":
					lists++
					var domains []map[string]any
					for id, name := range []string{"*", "api.example.com", "www.example.com", "*.example.com", "example.com", "nested.api.example.com", "http.example.com"} {
						protocol := "HTTPS"
						if id == 0 || id == 6 {
							protocol = "HTTP"
						}
						identifier, forceHttps := "old-cn-hangzhou", !tc.forceHttps
						if slices.Contains(updated, id) {
							identifier, protocol, forceHttps = "123-cn-hangzhou", "HTTPS", tc.forceHttps
						}
						// 回读分别延迟证书、协议和强制跳转，确保部署不会过早报告成功。
						if lists == 2 {
							switch tc.name {
							case "wildcard":
								identifier = "old-cn-hangzhou"
							case "HTTP domain":
								protocol = "HTTP"
							case "force HTTPS":
								forceHttps = false
							}
						}
						domains = append(domains, map[string]any{"Id": id, "Name": name, "Protocol": protocol, "CertIdentifier": identifier, "MustHttps": forceHttps})
					}
					writeResponse(w, domains)
				case "GetGatewayDomainDetail":
					id, _ := strconv.Atoi(r.Form.Get("Id"))
					if id == 6 {
						t.Error("HTTP domains have no existing HTTPS settings to read")
					}
					detail := map[string]any{"Id": id, "MustHttps": !tc.forceHttps}
					if id != 3 {
						for key, value := range settings {
							detail[key] = value
						}
						detail["MtlsEnabled"] = true
						delete(detail, "TlsCipherSuitesConfigJSON")
						detail["TlsCipherSuitesConfig"] = map[string]any{"ConfigType": "CUSTOM", "TlsCipherSuites": []string{"ECDHE-RSA-AES128-GCM-SHA256"}}
					}
					writeResponse(w, detail)
				case "UpdateGatewayDomain":
					id, err := strconv.Atoi(r.Form.Get("Id"))
					if err != nil || id <= 0 {
						t.Errorf("invalid domain ID: %q", r.Form.Get("Id"))
					}
					if r.Form.Get("CertIdentifier") != "123-cn-hangzhou" {
						t.Error("certificate identifier was not reused")
					}
					if r.Form.Get("Protocol") != "HTTPS" || r.Form.Get("MustHttps") != strconv.FormatBool(tc.forceHttps) {
						t.Error("HTTPS and forced redirects must follow the deployment configuration")
					}
					for key, value := range settings {
						if id == 6 || id == 3 {
							if r.Form.Has(key) {
								t.Errorf("absent setting %s must not be invented", key)
							}
						} else if r.Form.Get(key) != value {
							t.Errorf("setting %s = %q, want %q", key, r.Form.Get(key), value)
						}
					}
					updated = append(updated, id)
					writeResponse(w, id)
				default:
					t.Errorf("unexpected action: %v", r.Header.Get("x-acs-action"))
					http.Error(w, "unexpected action", http.StatusBadRequest)
				}
			})
			_, err := d.Deploy(t.Context(), certPEM, "test-key")
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("error = %v, want %q", err, tc.wantError)
				}
				if certmgr.uploads != 0 || len(updated) != 0 {
					t.Fatal("invalid target must not upload or deploy a certificate")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(updated, tc.wantIDs) || certmgr.uploads != 1 || lists < 2 {
				t.Fatalf("updated=%v, uploads=%d, lists=%d; want IDs=%v and one upload with readback", updated, certmgr.uploads, lists, tc.wantIDs)
			}
			if (tc.name == "HTTP domain" || tc.name == "wildcard" || tc.name == "force HTTPS") && lists < 3 {
				t.Fatal("must wait for HTTPS, forced redirects and the new certificate binding")
			}
		})
	}
}

func TestDeployFailures(t *testing.T) {
	certPEM := testCertificate(t)
	for _, protocol := range []string{"HTTP", "HTTPS"} {
		failures := []string{"list HTTP error", "list business error", "upload", "missing certificate ID", "update HTTP error", "update success false", "update missing result", "update invalid result", "readback", "timeout", "invalid certificate"}
		if protocol == "HTTPS" {
			failures = append(failures, "detail HTTP error", "detail business error", "detail missing", "detail mismatched")
		}
		for _, failure := range failures {
			t.Run(protocol+"/"+failure, func(t *testing.T) {
				t.Parallel()
				lists, updates := 0, 0
				d, certmgr := newTestDeployer(t, &DeployerConfig{GatewayId: "gw-test", DomainMatchPattern: "certsan"}, func(w http.ResponseWriter, r *http.Request) {
					_ = r.ParseForm()
					switch r.Header.Get("x-acs-action") {
					case "ListGatewayDomain":
						lists++
						if failure == "list HTTP error" {
							w.WriteHeader(http.StatusForbidden)
							_, _ = w.Write([]byte(`{"Code":"NoPermission","Message":"denied","RequestId":"test-request"}`))
							return
						}
						if failure == "list business error" || (failure == "readback" && lists > 1) {
							_, _ = w.Write([]byte(`{"Success":false,"Code":403,"Message":"denied","RequestId":"test-request"}`))
							return
						}
						writeResponse(w, []map[string]any{
							{"Id": 1, "Name": "api.example.com", "Protocol": protocol, "CertIdentifier": "old-cn-hangzhou"},
							{"Id": 2, "Name": "www.example.com", "Protocol": protocol, "CertIdentifier": "old-cn-hangzhou"},
						})
					case "GetGatewayDomainDetail":
						id, _ := strconv.Atoi(r.Form.Get("Id"))
						if id == 1 {
							switch failure {
							case "detail HTTP error":
								w.WriteHeader(http.StatusForbidden)
								_, _ = w.Write([]byte(`{"Code":"NoPermission","Message":"denied","RequestId":"test-request"}`))
								return
							case "detail business error":
								_, _ = w.Write([]byte(`{"Success":false,"Code":403,"Message":"denied","RequestId":"test-request"}`))
								return
							case "detail missing":
								writeResponse(w, nil)
								return
							case "detail mismatched":
								writeResponse(w, map[string]any{"Id": 99})
								return
							}
						}
						writeResponse(w, map[string]any{"Id": id})
					case "UpdateGatewayDomain":
						updates++
						id, _ := strconv.Atoi(r.Form.Get("Id"))
						if id == 1 {
							switch failure {
							case "update HTTP error":
								w.WriteHeader(http.StatusBadRequest)
								_, _ = w.Write([]byte(`{"Code":"InvalidParameter","Message":"failed","RequestId":"test-request"}`))
								return
							case "update success false":
								_ = json.NewEncoder(w).Encode(map[string]any{"Success": false, "Data": id, "Code": 400})
								return
							case "update missing result":
								writeResponse(w, nil)
								return
							case "update invalid result":
								writeResponse(w, 99)
								return
							}
						}
						writeResponse(w, id)
					default:
						t.Errorf("unexpected action %q", r.Header.Get("x-acs-action"))
					}
				})
				if failure == "upload" {
					certmgr.err = errors.New("upload failed")
				}
				if failure == "missing certificate ID" {
					certmgr.identifier = ""
				}
				ctx := t.Context()
				if failure == "timeout" {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, 2500*time.Millisecond)
					defer cancel()
				}
				input := certPEM
				if failure == "invalid certificate" {
					input = "invalid PEM"
				}
				_, err := d.Deploy(ctx, input, "test-key")
				if err == nil {
					t.Fatal("expected an error")
				}
				switch {
				case failure == "timeout":
					if !errors.Is(err, context.DeadlineExceeded) || lists < 2 {
						t.Fatalf("stale binding must time out after readback: lists=%d, error=%v", lists, err)
					}
				case failure == "readback":
					if !strings.Contains(err.Error(), "verify MSE") {
						t.Fatalf("unexpected readback error: %v", err)
					}
				case strings.HasPrefix(failure, "update"):
					if updates != 2 || !strings.Contains(err.Error(), `"api.example.com" (id 1)`) {
						t.Fatalf("partial failure must identify the failed domain and attempt remaining domains: updates=%d, error=%v", updates, err)
					}
				case strings.HasPrefix(failure, "detail"):
					if updates != 1 || !strings.Contains(err.Error(), `"api.example.com" (id 1)`) {
						t.Fatalf("failed detail lookup must prevent updating that domain: updates=%d, error=%v", updates, err)
					}
				default:
					if updates != 0 {
						t.Fatalf("unexpected deployment after validation, list or upload failure: %v", err)
					}
				}
			})
		}
	}
}

func TestDeployCancellation(t *testing.T) {
	d, certmgr := newTestDeployer(t, &DeployerConfig{GatewayId: "gw-test", DomainMatchPattern: "certsan"}, func(w http.ResponseWriter, r *http.Request) {
		t.Error("canceled deployment must not send API requests")
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := d.Deploy(ctx, testCertificate(t), "test-key"); err == nil {
		t.Fatal("expected cancellation error")
	}
	if certmgr.uploads != 0 {
		t.Fatal("canceled deployment uploaded a certificate")
	}
}
