package certacme_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMigratedDNSAccessOwnerWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_MIGRATED_DNS_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_MIGRATED_DNS_CORE to the core executable")
	}
	pluginDir := os.Getenv("CERTIMATE_MIGRATED_DNS_PLUGINS")
	plugins := stageMigratedDNSWorkflowPlugin(t, pluginDir, "acmehttpreq-dns", "acmehttpreq")
	plugins.cname = true
	plugins.configureDNS = configureACMEHTTPReqFixture
	runDNSWorkflow(t, binary, "acmehttpreq", plugins)
}

func configureACMEHTTPReqFixture(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
	const username, password = "fixture-user", "fixture-password"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		user, secret, ok := r.BasicAuth()
		if !ok || user != username || secret != password {
			t.Error("challenge request did not use the saved Basic Auth credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if f.mode == "auth-failure" {
			f.denied++
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var record struct {
			FQDN  string `json:"fqdn"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil || record.FQDN != f.challengeRecordName() || record.Value == "" {
			t.Errorf("expected delegated TXT record, got %+v (%v)", record, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /present":
			f.value = record.Value
			f.creates++
		case "POST /cleanup":
			if f.value != record.Value {
				t.Error("cleanup did not target the TXT record that was created")
			}
			f.value = ""
			f.deletes++
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return b3DNSProviderConfig{
		Provider: "acmehttpreq",
		Access:   map[string]any{"endpoint": server.URL, "username": username, "password": password},
		Apply:    map[string]any{"disableFollowCNAME": false},
	}
}
