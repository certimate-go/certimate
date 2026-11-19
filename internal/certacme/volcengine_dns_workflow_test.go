package certacme_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestVolcengineDNSWorkflow(t *testing.T) {
	coreBinary := os.Getenv("CERTIMATE_VOLCENGINE_DNS_CORE")
	pluginDir := os.Getenv("CERTIMATE_VOLCENGINE_DNS_PLUGINS")
	if coreBinary == "" || pluginDir == "" {
		t.Skip("set CERTIMATE_VOLCENGINE_DNS_CORE and CERTIMATE_VOLCENGINE_DNS_PLUGINS")
	}
	for _, provider := range []string{"volcengine", "volcengine-dns"} {
		t.Run(provider, func(t *testing.T) {
			runDNSWorkflow(t, coreBinary, "volcengine", dnsWorkflowPlugins{
				dir:              pluginDir,
				accessByProvider: map[string]string{"volcengine-dns": "volcengine"},
				applyPlugin:      "volcengine-dns.exe",
				pollingInterval:  1,
				wildcardOrder:    true,
				cname:            true,
				configureDNS: func(t *testing.T, fixture *b3DNSFixture) b3DNSProviderConfig {
					return configureVolcengineDNSFixture(t, fixture, provider)
				},
			})
		})
	}
}

func configureVolcengineDNSFixture(t *testing.T, f *b3DNSFixture, provider string) b3DNSProviderConfig {
	const (
		accessKey = "fixture-workflow-key"
		secretKey = "fixture-workflow-secret"
		zone      = "example.test"
		zoneID    = 17
		ttl       = 600
		existing  = "already-present"
	)
	f.unrelatedTXT = "pre-existing-value"
	records := map[string]string{existing: f.unrelatedTXT}
	expectedHost := strings.TrimSuffix(f.challengeRecordName(), "."+zone+".")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query()
		if !strings.Contains(r.Header.Get("Authorization"), "Credential="+accessKey+"/") {
			t.Error("DNS API request did not use the saved Volcano Engine access record")
			http.Error(w, "invalid credentials", http.StatusForbidden)
			return
		}
		switch query.Get("Action") {
		case "ListZones":
			if query.Get("SearchMode") != "exact" {
				t.Error("Volcano Engine DNS zone lookup must be exact")
			}
			if query.Get("Key") == zone {
				fmt.Fprintf(w, `{"Result":{"Total":1,"Zones":[{"ZID":%d,"ZoneName":%q}]}}`, zoneID, zone)
			} else {
				fmt.Fprint(w, `{"Result":{"Total":0,"Zones":[]}}`)
			}
		case "CreateRecord":
			var record struct {
				Host  string
				Type  string
				Value string
				TTL   int
				ZID   int
			}
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil ||
				record.Type != "TXT" || record.Host != expectedHost || record.ZID != zoneID || record.TTL != ttl {
				t.Errorf("expected delegated TXT with workflow TTL, got %+v (%v)", record, err)
				http.Error(w, "invalid record", http.StatusBadRequest)
				return
			}
			if f.mode == "auth-failure" {
				f.denied++
				fmt.Fprint(w, `{"ResponseMetadata":{"Error":{"Code":"Forbidden","Message":"denied"}}}`)
				return
			}
			if rejectChallengeWrite(f, record.Value) {
				fmt.Fprint(w, `{"ResponseMetadata":{"Error":{"Code":"Forbidden","Message":"wildcard write denied"}}}`)
				return
			}
			applyExternalDNSUpdate(t, f, f.challengeRecordName(), record.Value, true)
			id := fmt.Sprintf("record-%d", f.creates)
			records[id] = record.Value
			fmt.Fprintf(w, `{"Result":{"RecordID":%q}}`, id)
		case "DeleteRecord":
			var record struct{ RecordID string }
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				t.Errorf("decode deletion request: %v", err)
				http.Error(w, "invalid record ID", http.StatusBadRequest)
				return
			}
			value, created := records[record.RecordID]
			if !created || record.RecordID == existing {
				t.Errorf("unexpected deletion of record %q", record.RecordID)
				http.Error(w, "unexpected record ID", http.StatusBadRequest)
				return
			}
			applyExternalDNSUpdate(t, f, f.challengeRecordName(), value, false)
			delete(records, record.RecordID)
			fmt.Fprint(w, `{"Result":{}}`)
		default:
			t.Errorf("unexpected Volcano Engine DNS action %q", query.Get("Action"))
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return b3DNSProviderConfig{
		Provider: provider,
		Access:   map[string]any{"accessKeyId": accessKey, "secretAccessKey": secretKey},
		Apply:    map[string]any{"dnsTTL": ttl, "disableFollowCNAME": false},
		Env:      []string{"VOLC_HOST=" + strings.TrimPrefix(server.URL, "http://"), "VOLC_SCHEME=http"},
	}
}
