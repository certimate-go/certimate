package engine

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/certimate-go/certimate/internal/certapply"
	"github.com/certimate-go/certimate/internal/domain"
)

type nativeDNSAccessRepo struct {
	nativeTestAccessRepo
	records   map[string]*domain.Access
	requested []string
}

func (r *nativeDNSAccessRepo) GetById(_ context.Context, id string) (*domain.Access, error) {
	r.requested = append(r.requested, id)
	if access, ok := r.records[id]; ok {
		return access, nil
	}
	return nil, domain.ErrRecordNotFound
}

func TestNativeTencentCloudDNSCredentials(t *testing.T) {
	for _, provider := range []string{"tencentcloud-dns", "aliyun-dns", "cloudflare"} {
		t.Run(provider, func(t *testing.T) {
			config := nativeTestNodeConfig()
			config["challengeType"], config["provider"], config["providerAccessId"] = "dns-01", provider, "dns-access"
			config["dnsPropagationWait"], config["dnsTTL"] = 12, 600
			repo := &nativeDNSAccessRepo{records: map[string]*domain.Access{
				"tc-access":  {Provider: "tencentcloud", Config: map[string]any{"secretId": "ssl-id", "secretKey": "ssl-key"}},
				"dns-access": {Provider: strings.TrimSuffix(provider, "-dns"), Config: map[string]any{"secretId": "dns-id", "secretKey": "dns-key", "accessKeyId": "dns-id", "accessKeySecret": "dns-key", "apiToken": "dns-token"}},
			}}
			executor, outputs, _ := nativeTestExecutor(t, nil)
			executor.accessRepo = repo
			calls := 0
			executor.newTencentCloudClient = func(cfg certapply.TencentCloudConfig, _ *slog.Logger) (tencentCloudIssuer, error) {
				if cfg.SecretId != "ssl-id" || cfg.SecretKey != "ssl-key" || cfg.DNSProvider == nil {
					t.Fatal("SSL and DNS credentials were mixed")
				}
				return nativeTestIssuer(func(_ context.Context, req certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
					calls++
					if req.DvAuthMethod != "DNS" || req.DNSPropagationWait != 12 {
						t.Fatalf("wrong DNS request: %v", req)
					}
					if calls == 1 {
						if err := created("dns-order"); err != nil {
							return nil, err
						}
					} else if req.CertificateId != "dns-order" {
						t.Fatal("DNS retry lost pending order")
					}
					return nil, context.DeadlineExceeded
				}), nil
			}
			for _, run := range []string{"run-1", "run-2"} {
				if _, err := executor.Execute(nativeTestContext(config, run)); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("got %v", err)
				}
			}
			if outputs.saves != 2 || !reflect.DeepEqual(repo.requested, []string{"tc-access", "dns-access", "tc-access", "dns-access"}) {
				t.Fatalf("wrong persistence or credential reads: %v", repo.requested)
			}
		})
	}
}

func TestNativeTencentCloudDNSValidation(t *testing.T) {
	for _, change := range []domain.WorkflowNodeConfig{
		{"provider": "unsupported"}, {"providerAccessId": ""}, {"dnsTTL": -1}, {"dnsTTL": 86401}, {"dnsPropagationWait": 3601},
	} {
		config := nativeTestNodeConfig()
		config["challengeType"], config["provider"], config["providerAccessId"] = "dns-01", "tencentcloud-dns", "dns-access"
		maps.Copy(config, change)
		executor, _, _ := nativeTestExecutor(t, nil)
		executor.newTencentCloudClient = func(certapply.TencentCloudConfig, *slog.Logger) (tencentCloudIssuer, error) {
			t.Fatal("invalid DNS config reached issuer")
			return nil, nil
		}
		if _, err := executor.Execute(nativeTestContext(config, "run")); err == nil {
			t.Fatalf("accepted %v", change)
		}
	}
	for _, access := range []*domain.Access{
		{Provider: "aliyun"}, {Provider: "tencentcloud", Reserve: "ca"}, {Provider: "tencentcloud", Config: map[string]any{}},
	} {
		config := nativeTestNodeConfig()
		config["challengeType"], config["provider"], config["providerAccessId"] = "dns-01", "tencentcloud-dns", "dns-access"
		executor, _, _ := nativeTestExecutor(t, nil)
		executor.accessRepo = &nativeDNSAccessRepo{records: map[string]*domain.Access{
			"tc-access": {Provider: "tencentcloud", Config: map[string]any{"secretId": "ssl-id", "secretKey": "ssl-key"}}, "dns-access": access,
		}}
		executor.newTencentCloudClient = func(certapply.TencentCloudConfig, *slog.Logger) (tencentCloudIssuer, error) {
			t.Fatal("invalid access reached issuer")
			return nil, nil
		}
		if _, err := executor.Execute(nativeTestContext(config, "run")); err == nil {
			t.Fatal("accepted invalid DNS access")
		}
	}
}

func TestNativeTencentCloudDNSPendingMode(t *testing.T) {
	config := nativeTestNodeConfig()
	last := &domain.WorkflowOutput{NodeConfig: config, Outputs: []*domain.WorkflowOutputEntry{{Name: tencentCloudPendingCertificateKey, Value: "pending"}}}
	changed := maps.Clone(config)
	changed["challengeType"] = "dns-auto"
	current := changed.AsBizApply()
	if got := tencentCloudPendingCertificate(&current, last); got != "pending" {
		t.Fatal("legacy automatic order was not reused")
	}
	changed["challengeType"] = "dns-01"
	current = changed.AsBizApply()
	if got := tencentCloudPendingCertificate(&current, last); got != "" {
		t.Fatal("reused order with incompatible verification mode")
	}
	last.NodeConfig = maps.Clone(changed)
	changed["provider"], changed["providerAccessId"] = "cloudflare", "another-dns-access"
	current = changed.AsBizApply()
	if got := tencentCloudPendingCertificate(&current, last); got != "pending" {
		t.Fatal("changing DNS credentials unnecessarily replaced the SSL order")
	}
}
