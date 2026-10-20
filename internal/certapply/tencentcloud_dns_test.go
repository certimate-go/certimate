package certapply

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type nativeDNSProviderFunc func(context.Context, DNSRecord) (func(context.Context) error, error)

func (f nativeDNSProviderFunc) Present(ctx context.Context, record DNSRecord) (func(context.Context) error, error) {
	return f(ctx, record)
}

func TestTencentCloudDNSRecords(t *testing.T) {
	for _, tc := range []struct {
		name, key, domain, recordType, want string
		invalid                             bool
	}{
		{name: "relative", key: "_dnsauth.test", domain: "example.com", want: "_dnsauth.test.example.com"},
		{name: "full key", key: "_dnsauth.test.example.com.", domain: "example.com", want: "_dnsauth.test.example.com"},
		{name: "full domain", key: "_dnsauth", domain: "_dnsauth.test.example.com", want: "_dnsauth.test.example.com"},
		{name: "subdomain", key: "_dnsauth", domain: "test.example.com", want: "_dnsauth.test.example.com"},
		{name: "TXT", key: "_dnsauth", domain: "test.example.com", recordType: "TXT", want: "_dnsauth.test.example.com"},
		{name: "unrelated", key: "_dnsauth", domain: "other.example.net", invalid: true},
		{name: "suffix boundary", key: "_dnsauth", domain: "notexample.com", invalid: true},
		{name: "CNAME", key: "_dnsauth", domain: "test.example.com", recordType: "CNAME", invalid: true},
		{name: "bad name", key: "_dnsauth/", domain: "test.example.com", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth := &ssl.DvAuths{DvAuthKey: common.StringPtr(tc.key), DvAuthDomain: common.StringPtr(tc.domain), DvAuthValue: common.StringPtr("raw-validation-value"), DvAuthVerifyType: common.StringPtr(tc.recordType)}
			records, err := tencentCloudDNSRecords(&ssl.DvAuthDetail{DvAuths: []*ssl.DvAuths{auth, auth}}, "test.example.com")
			if tc.invalid {
				if err == nil {
					t.Fatal("accepted invalid validation record")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(records, []DNSRecord{{Name: tc.want, Value: "raw-validation-value"}}) {
				t.Fatalf("got %v, %v", records, err)
			}
		})
	}
	for _, detail := range []*ssl.DvAuthDetail{nil, {}, {DvAuthKey: common.StringPtr("_dnsauth")}} {
		if records, err := tencentCloudDNSRecords(detail, "test.example.com"); err != nil || len(records) != 0 {
			t.Fatalf("incomplete verification data should wait: %v, %v", records, err)
		}
	}
}

func TestTencentCloudDNSObtain(t *testing.T) {
	cert, key := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(24*time.Hour))
	for _, resumedStatus := range []int{-1, 0, 4, 13} {
		t.Run(string(rune('a'+resumedStatus+1)), func(t *testing.T) {
			var events []string
			persisted := resumedStatus >= 0
			queries := 0
			client, logs := newTestTencentCloudClient(t, func(action string, body map[string]any) map[string]any {
				events = append(events, action)
				switch action {
				case "ApplyCertificate":
					if resumedStatus >= 0 {
						t.Fatal("resumed order applied again")
					}
					if body["DvAuthMethod"] != "DNS" || body["DeleteDnsAutoRecord"] != nil {
						t.Fatal("wrong DNS mode parameters")
					}
					return map[string]any{"CertificateId": "dns-order"}
				case "DescribeCertificateDetail":
					if !persisted {
						t.Fatal("record must be persisted before DNS operations")
					}
					queries++
					if queries == 1 {
						return map[string]any{"Status": 0}
					}
					if queries > 3 {
						return map[string]any{"Status": 1}
					}
					status := resumedStatus
					if status < 0 {
						status = 0
					}
					return map[string]any{"Status": status, "VerifyType": "DNS", "DvAuthDetail": map[string]any{"DvAuthKey": "_dnsauth", "DvAuthDomain": "test.example.com", "DvAuthValue": "not-an-acme-hash"}}
				case "CompleteCertificate":
					if queries < 2 {
						t.Fatal("completed without verification records")
					}
					return map[string]any{}
				case "DownloadCertificate":
					return map[string]any{"Content": testCertificateArchive(t, map[string]string{"Nginx/test_bundle.crt": cert, "Nginx/test.key": key})}
				default:
					t.Fatalf("unexpected action: %s", action)
					return nil
				}
			})
			client.dnsProvider = nativeDNSProviderFunc(func(ctx context.Context, record DNSRecord) (func(context.Context) error, error) {
				if !persisted || record.Name != "_dnsauth.test.example.com" || record.Value != "not-an-acme-hash" {
					t.Fatalf("invalid DNS operation: %v", record)
				}
				events = append(events, "Present")
				return func(ctx context.Context) error {
					if ctx.Err() != nil {
						t.Fatal("cleanup context canceled")
					}
					events = append(events, "CleanUp")
					return nil
				}, nil
			})
			request := validTencentCloudRequest()
			request.KeyAlgorithm, request.DvAuthMethod = "EC256", "DNS"
			if resumedStatus >= 0 {
				request.CertificateId = "dns-order"
			}
			result, err := client.Obtain(context.Background(), request, func(string) error { persisted = true; return nil })
			if err != nil || result == nil || result.Certificate != cert {
				t.Fatalf("issuance failed: %v", err)
			}
			counts := map[string]int{}
			for _, event := range events {
				counts[event]++
			}
			if counts["Present"] != 1 || counts["CleanUp"] != 1 || events[len(events)-1] != "CleanUp" {
				t.Fatalf("unexpected lifecycle: %v", events)
			}
			if resumedStatus == 4 && counts["CompleteCertificate"] != 0 {
				t.Fatal("completed an already processing order")
			}
			if strings.Contains(logs.String(), "not-an-acme-hash") {
				t.Fatal("validation value leaked to logs")
			}
		})
	}
}

func TestTencentCloudDNSFailureCleanup(t *testing.T) {
	for _, failure := range []string{"present", "complete", "cancel", "cleanup"} {
		t.Run(failure, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cleaned := false
			client, logs := newTestTencentCloudClient(t, func(action string, _ map[string]any) map[string]any {
				switch action {
				case "DescribeCertificateDetail":
					return map[string]any{"Status": 0, "DvAuthDetail": map[string]any{"DvAuthKey": "_dnsauth", "DvAuthDomain": "test.example.com", "DvAuthValue": "token"}}
				case "CompleteCertificate":
					return map[string]any{"Error": map[string]any{"Code": "FailedOperation", "Message": "validation failed"}}
				default:
					t.Fatalf("unexpected action %s", action)
					return nil
				}
			})
			client.dnsProvider = nativeDNSProviderFunc(func(context.Context, DNSRecord) (func(context.Context) error, error) {
				cleanup := func(cleanupCtx context.Context) error {
					if cleanupCtx.Err() != nil {
						t.Fatal("cleanup must survive cancellation")
					}
					if _, ok := cleanupCtx.Deadline(); !ok {
						t.Fatal("cleanup has no deadline")
					}
					cleaned = true
					if failure == "cleanup" {
						return errors.New("cleanup denied")
					}
					return nil
				}
				if failure == "present" {
					return cleanup, errors.New("write failed")
				}
				if failure == "cancel" {
					cancel()
				}
				return cleanup, nil
			})
			request := validTencentCloudRequest()
			request.CertificateId, request.DvAuthMethod, request.DNSPropagationWait = "pending-dns", "DNS", 0
			if failure == "cancel" {
				request.DNSPropagationWait = 30
			}
			_, err := client.Obtain(ctx, request, nil)
			if err == nil || !cleaned {
				t.Fatalf("expected error and cleanup, got %v, %v", err, cleaned)
			}
			if failure == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
			if failure == "cleanup" && !strings.Contains(logs.String(), "failed to clean up") {
				t.Fatal("cleanup failure was not reported")
			}
		})
	}
}

func TestTencentCloudDNSProviderIntegration(t *testing.T) {
	cert, key := testCertificatePEM(t, "EC256", "test.example.com", time.Now().Add(24*time.Hour))
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		t.Run(provider, func(t *testing.T) {
			dns := &dnsTestCloud{t: t, provider: provider, zone: "test.example.com", rr: "_dnsauth", record: DNSRecord{Name: "_dnsauth.test.example.com", Value: "raw-tencentcloud-token"}}
			completed := false
			client, _ := newTestTencentCloudClient(t, func(action string, body map[string]any) map[string]any {
				switch action {
				case "ApplyCertificate":
					if body["DvAuthMethod"] != "DNS" {
						t.Fatal("expected DNS mode")
					}
					return map[string]any{"CertificateId": "integration-order"}
				case "DescribeCertificateDetail":
					if completed {
						return map[string]any{"Status": 1}
					}
					return map[string]any{"Status": 0, "DvAuthDetail": map[string]any{"DvAuthKey": "_dnsauth", "DvAuthDomain": "test.example.com", "DvAuthValue": dns.record.Value}}
				case "CompleteCertificate":
					if dns.creates != 1 {
						t.Fatal("validation started before DNS record creation")
					}
					completed = true
					return map[string]any{}
				case "DownloadCertificate":
					return map[string]any{"Content": testCertificateArchive(t, map[string]string{"Nginx/test_bundle.crt": cert, "Nginx/test.key": key})}
				default:
					t.Fatalf("unexpected SSL action %s", action)
					return nil
				}
			})
			client.dnsProvider = dns.providerClient()
			request := validTencentCloudRequest()
			request.DvAuthMethod, request.KeyAlgorithm = "DNS", "EC256"
			result, err := client.Obtain(context.Background(), request, func(string) error {
				if dns.creates != 0 || len(dns.deletes) != 0 {
					t.Fatal("DNS records changed before order persistence")
				}
				return nil
			})
			if err != nil || result == nil || result.Certificate != cert || len(dns.deletes) != 1 {
				t.Fatalf("SSL and DNS integration failed: %v, cleanup=%v", err, dns.deletes)
			}
		})
	}
}

func TestTencentCloudDNSRequiresProvider(t *testing.T) {
	client, _ := newTestTencentCloudClient(t, func(string, map[string]any) map[string]any { t.Fatal("invalid request reached cloud"); return nil })
	request := validTencentCloudRequest()
	request.DvAuthMethod = "DNS"
	if _, err := client.Obtain(context.Background(), request, nil); err == nil {
		t.Fatal("missing DNS provider accepted")
	}
}

func TestTencentCloudDNSPreflightRejectsWrongAccount(t *testing.T) {
	for _, provider := range []string{"tencentcloud", "aliyun", "cloudflare"} {
		for _, failure := range []string{"no-zone", "permission", "canceled"} {
			t.Run(provider+"/"+failure, func(t *testing.T) {
				dns := dnsTestFixture(t, provider)
				dns.boundaryOnly = failure == "no-zone"
				if failure == "permission" {
					dns.fail = "zones"
				}
				client, _ := newTestTencentCloudClient(t, func(string, map[string]any) map[string]any {
					t.Fatal("invalid DNS account must not consume a free certificate order")
					return nil
				})
				client.dnsProvider = dns.providerClient()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if failure == "canceled" {
					cancel()
				}
				request := validTencentCloudRequest()
				request.DvAuthMethod = "DNS"
				result, err := client.Obtain(ctx, request, func(string) error {
					t.Fatal("invalid DNS account must not persist a new order")
					return nil
				})
				if result != nil || err == nil || dns.creates != 0 || len(dns.deletes) != 0 {
					t.Fatalf("expected read-only preflight failure: result=%v err=%v", result, err)
				}
				if failure == "no-zone" && (!strings.Contains(err.Error(), request.Domains[0]) || !strings.Contains(err.Error(), "selected DNS account")) {
					t.Fatalf("missing actionable domain/account diagnostic: %v", err)
				}
				if failure == "canceled" && !errors.Is(err, context.Canceled) {
					t.Fatalf("lost context cancellation: %v", err)
				}
				if strings.Contains(err.Error(), dnsTestSecret) {
					t.Fatal("DNS credentials leaked in preflight error")
				}
			})
		}
	}
}
