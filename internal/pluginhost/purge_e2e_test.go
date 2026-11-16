package pluginhost_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt"
	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/certmgmt/purgers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/plugin"
)

const (
	purgePluginProvider = "baiducloud-cert-purger"
	purgeLegacyProvider = string(domain.PurgeProviderTypeBaiduCloudCert)
	purgeAccessType     = string(domain.AccessProviderTypeBaiduCloud)
)

type purgeVendorCall struct {
	ExpirySeconds int64          `json:"expirySeconds"`
	Version       string         `json:"version"`
	Auth          string         `json:"-"`
	Config        map[string]any `json:"config"`
}

type purgeVendorService struct {
	mu     sync.Mutex
	calls  []purgeVendorCall
	count  int
	status int
}

func (s *purgeVendorService) record(call purgeVendorCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, call)
}

func (s *purgeVendorService) total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *purgeVendorService) snapshot() []purgeVendorCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]purgeVendorCall, len(s.calls))
	copy(out, s.calls)
	return out
}

func newPurgeVendorServer(t *testing.T, service *purgeVendorService) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var call purgeVendorCall
		if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
			t.Errorf("vendor payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		call.Auth = r.Header.Get("Authorization")
		service.record(call)
		service.mu.Lock()
		status, count := service.status, service.count
		service.mu.Unlock()
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"count": count})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func setupPurgeE2E(t *testing.T, fixture map[string]any) (*pluginhost.Catalog, plugin.PluginConfig) {
	t.Helper()
	root := t.TempDir()
	if _, ok := fixture["omitAccessSchema"]; !ok {
		fixture["omitAccessSchema"] = true
	}
	if _, ok := fixture["capability"]; !ok {
		fixture["capability"] = plugin.CapabilityPurge
	}
	setupE2EPlugin(t, root, purgePluginProvider, purgeAccessType, fixture)
	cfg := plugin.PluginConfig{PluginDir: root, CoreVersion: "0.4.28"}
	catalog, errs := pluginhost.ScanAndRegister(context.Background(), cfg, slog.Default())
	if len(errs) != 0 {
		t.Fatalf("purge plugin registration errors: %+v", errs)
	}
	t.Cleanup(catalog.Close)
	return catalog, cfg
}

func purgeViaClient(t *testing.T, provider string, vendorURL string, expiry time.Duration) (*certmgmt.PurgeCertificateResponse, error) {
	t.Helper()
	client := certmgmt.NewClient()
	return client.PurgeCertificate(context.Background(), &certmgmt.PurgeCertificateRequest{
		Provider:               domain.PurgeProviderType(provider),
		ProviderAccessConfig:   map[string]any{"url": vendorURL, "token": "purge-secret"},
		ProviderExtendedConfig: map[string]any{"resourceGroupId": "rg-1"},
		Expiry:                 expiry,
	})
}

func TestPurgeE2E_LegacyAndCanonicalIDsReachPlugin(t *testing.T) {
	vendor := &purgeVendorService{count: 3}
	srv := newPurgeVendorServer(t, vendor)

	catalog, _ := setupPurgeE2E(t, map[string]any{"schemaRevision": "v1"})
	entries := catalog.Entries()
	if len(entries) != 1 || entries[0].Capability != plugin.CapabilityPurge {
		t.Fatalf("purge plugin catalog entry wrong: %+v", entries)
	}
	if entries[0].AccessProviderType != purgeAccessType || entries[0].OwnsAccess {
		t.Fatalf("purge plugin must reuse the %q built-in access without owning it: %+v", purgeAccessType, entries[0])
	}

	if _, err := purgers.Registries.Get(domain.PurgeProviderType(purgePluginProvider)); err != nil {
		t.Fatalf("purge plugin not in purger registry: %v", err)
	}

	for _, provider := range []string{purgePluginProvider, purgeLegacyProvider} {
		res, err := purgeViaClient(t, provider, srv.URL, 7*24*time.Hour)
		if err != nil {
			t.Fatalf("purge via %q: %v", provider, err)
		}
		if res.Count != 3 {
			t.Fatalf("purge via %q returned count %d, want 3", provider, res.Count)
		}
	}

	calls := vendor.snapshot()
	if len(calls) != 2 {
		t.Fatalf("vendor saw %d purge calls, want 2", len(calls))
	}
	for _, call := range calls {
		if call.ExpirySeconds != 7*24*3600 {
			t.Fatalf("vendor expiry seconds = %d, want %d", call.ExpirySeconds, 7*24*3600)
		}
		if call.Auth != "Bearer purge-secret" {
			t.Fatalf("shared access credentials not forwarded: %q", call.Auth)
		}
		if call.Config["resourceGroupId"] != "rg-1" {
			t.Fatalf("extended config not forwarded: %+v", call.Config)
		}
	}

	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(purgePluginProvider)); err == nil {
		t.Fatal("purge plugin must not register as deployer")
	}

	if env, ok := providerschema.Envelopes.GetEnvelope(purgeAccessType, providerschema.CategoryAccess); ok && env != nil {
		t.Fatalf("purge plugin must not ship a second access schema for built-in %q", purgeAccessType)
	}
	if env, err := (providerschema.NewProviderSchemaService(repository.NewProviderSchemaRepository())).GetByProviderType(context.Background(), purgePluginProvider, providerschema.CategoryPurge); err != nil || env == nil {
		t.Fatalf("purge schema envelope missing: %+v %v", env, err)
	}
}

func TestPurgeE2E_InvalidExpiryFailsBeforeVendor(t *testing.T) {
	vendor := &purgeVendorService{count: 3}
	srv := newPurgeVendorServer(t, vendor)

	setupPurgeE2E(t, map[string]any{"schemaRevision": "v1"})

	if _, err := purgeViaClient(t, purgePluginProvider, srv.URL, -time.Second); err == nil {
		t.Fatal("negative expiry must fail")
	}
	if got := vendor.total(); got != 0 {
		t.Fatalf("vendor received %d calls after invalid expiry, want 0", got)
	}
}

func TestPurgeE2E_MissingPluginReportsInstallableID(t *testing.T) {
	vendor := &purgeVendorService{}
	srv := newPurgeVendorServer(t, vendor)

	catalog, _ := setupPurgeE2E(t, map[string]any{"schemaRevision": "v1"})
	if _, err := purgeViaClient(t, purgePluginProvider, srv.URL, 0); err != nil {
		t.Fatalf("pre-uninstall purge: %v", err)
	}

	catalog.Close()
	if _, err := purgers.Registries.Get(domain.PurgeProviderType(purgePluginProvider)); err == nil {
		t.Fatal("uninstalled plugin remains in purger registry")
	}
	for _, provider := range []string{purgePluginProvider, purgeLegacyProvider} {
		_, err := purgeViaClient(t, provider, srv.URL, 0)
		if err == nil {
			t.Fatalf("purge via %q after uninstall must fail", provider)
		}
		if !strings.Contains(err.Error(), purgePluginProvider) {
			t.Fatalf("missing-plugin error must name %q: %v", purgePluginProvider, err)
		}
	}
	if got := vendor.total(); got != 1 {
		t.Fatalf("vendor saw %d calls, want only the pre-uninstall one", got)
	}
}

func TestPurgeE2E_DeployerWithSameNameUnaffected(t *testing.T) {
	setupPurgeE2E(t, map[string]any{"schemaRevision": "v1"})

	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(purgeLegacyProvider)); err == nil {
		t.Fatalf("%q must not silently become a deployer through the purge alias", purgeLegacyProvider)
	}
	factory, err := purgers.Registries.Get(domain.ResolvePurgeProviderAlias(domain.PurgeProviderType(purgeLegacyProvider)))
	if err != nil {
		t.Fatalf("legacy purge id must resolve to the plugin: %v", err)
	}
	if factory == nil {
		t.Fatal("nil purge factory")
	}
}
