package certacme_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

type persistRecord struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type persistPreparation struct {
	AccountID  string          `json:"accountId"`
	AccountURL string          `json:"accountUrl"`
	Records    []persistRecord `json:"records"`
}

type persistOrder struct {
	domains    []string
	authorized []bool
	submitted  []bool
	chain      []byte
}

type persistCAFixture struct {
	mu            sync.Mutex
	t             *testing.T
	base          string
	omitAccount   bool
	unsupported   bool
	cached        bool
	missingIssuer bool
	reject        bool
	dnsFailure    bool
	accounts      int
	issued        int
	checks        int
	polls         int
	queries       int
	writes        int
	orders        []*persistOrder
	records       map[string][]string
	issuer        *x509.Certificate
	key           *ecdsa.PrivateKey
}

func TestDNSPersistWorkflowE2E(t *testing.T) {
	binary := os.Getenv("CERTIMATE_DNS_PERSIST_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_DNS_PERSIST_CORE to a native core executable")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []struct {
		name        string
		omitAccount bool
		multiproc   bool
	}{
		{name: "litessl-shaped-multiprocess", multiproc: true},
		{name: "gts-shaped-multiprocess", omitAccount: true, multiproc: true},
		{name: "gts-shaped-single-process", omitAccount: true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			runPersistWorkflow(t, binary, mode.omitAccount, mode.multiproc)
		})
	}
}

func runPersistWorkflow(t *testing.T, binary string, omitAccount, multiproc bool) {
	f := &persistCAFixture{t: t, omitAccount: omitAccount, records: map[string][]string{}}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	f.key = key
	issuer := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Persistent DNS fixture CA"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, issuer, issuer, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	f.issuer, err = x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start ACME fixture listener: %v", err)
	}
	f.base = "http://" + listener.Addr().String()
	ca := &http.Server{Handler: http.HandlerFunc(f.acme)}
	go func() {
		if err := ca.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	defer ca.Close()
	udp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start DNS fixture listener: %v", err)
	}
	dnsServer := &dns.Server{PacketConn: udp, Handler: dns.HandlerFunc(f.dns)}
	go dnsServer.ActivateAndServe()
	defer dnsServer.Shutdown()
	api := startPersistCore(t, binary, multiproc)
	caID := api.create(t, "access", map[string]any{"name": "Persistent fixture CA", "provider": "acmeca", "reserve": "ca", "config": map[string]string{"endpoint": f.base + "/directory"}})
	domains := []string{"persist.example.test", "*.persist.example.test"}
	nameservers := []string{udp.LocalAddr().String()}
	payload := map[string]any{"domains": domains, "contactEmail": "persist@example.test", "caProvider": "acmeca", "caProviderAccessId": caID, "keyAlgorithm": "EC256", "nameservers": nameservers, "acmeProfile": "fixture"}
	var prepared persistPreparation
	if err := api.call("POST", "/api/acme-accounts/dns-persist/prepare", payload, &prepared); err != nil {
		t.Fatal(err)
	}
	accountURL := f.base + "/account/1"
	value := "fixture.example; accounturi=" + accountURL + "; policy=wildcard"
	name := "_validation-persist.persist.example.test."
	if prepared.AccountID == "" || prepared.AccountURL != accountURL || len(prepared.Records) != 1 {
		t.Fatalf("unexpected prepared account or apex/wildcard merge: %+v", prepared)
	}
	if record := prepared.Records[0]; dns.Fqdn(record.Name) != name || record.Type != "TXT" || record.Value != value {
		t.Fatalf("unexpected persistent TXT instruction: %+v", record)
	}
	var again persistPreparation
	if err := api.call("POST", "/api/acme-accounts/dns-persist/prepare", payload, &again); err != nil {
		t.Fatal(err)
	}
	if again.AccountID != prepared.AccountID || again.AccountURL != prepared.AccountURL || len(again.Records) != 1 || again.Records[0].Value != value {
		t.Fatalf("preparing again changed the account or TXT: %+v", again)
	}
	config := map[string]any{"domains": strings.Join(domains, ";"), "contactEmail": "persist@example.test", "challengeType": "dns-persist-01", "dnsPersistAccountId": prepared.AccountID, "caProvider": "acmeca", "caProviderAccessId": caID, "keyAlgorithm": "EC256", "nameservers": nameservers[0], "acmeProfile": "fixture", "disableARI": true, "skipBeforeExpiryDays": 30}
	workflowID := api.create(t, "workflow", persistWorkflow("Persistent DNS E2E", config))
	check := func(t *testing.T, wantReady bool, wantStatus string) {
		t.Helper()
		f.mu.Lock()
		beforeOrders, beforeAccounts := len(f.orders), f.accounts
		f.mu.Unlock()
		var result struct {
			Ready   bool            `json:"ready"`
			Records []persistRecord `json:"records"`
		}
		if err := api.call("POST", "/api/acme-accounts/dns-persist/check", map[string]any{"records": prepared.Records, "nameservers": nameservers}, &result); err != nil {
			t.Fatal(err)
		}
		if result.Ready != wantReady || len(result.Records) != 1 || result.Records[0].Status != wantStatus {
			t.Fatalf("DNS readiness = %+v, want ready=%v status=%s", result, wantReady, wantStatus)
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		if len(f.orders) != beforeOrders || f.accounts != beforeAccounts {
			t.Fatal("DNS check created an ACME order or account")
		}
	}
	for _, step := range []struct {
		name       string
		value      string
		ready      bool
		status     string
		caReject   bool
		dnsFailure bool
	}{
		{name: "missing", status: "missing"},
		{name: "first-issuance", value: value, ready: true, status: "ready"},
		{name: "repeat-issuance", value: value, ready: true, status: "ready"},
		{name: "wrong-account", value: strings.Replace(value, "/account/1", "/account/999", 1), status: "mismatch"},
		{name: "expired", value: value + "; persistUntil=1", status: "mismatch"},
		{name: "missing-wildcard-policy", value: strings.TrimSuffix(value, "; policy=wildcard"), status: "mismatch"},
		{name: "resolver-failure", value: value, status: "lookup_failed", dnsFailure: true},
		{name: "ca-rejected", value: value, ready: true, status: "ready", caReject: true},
		{name: "recovery", value: value, ready: true, status: "ready"},
	} {
		t.Run(step.name, func(t *testing.T) {
			f.mu.Lock()
			f.reject = step.caReject
			f.dnsFailure = step.dnsFailure
			f.records[name] = nil
			if step.value != "" {
				f.records[name] = []string{step.value}
			}
			beforeIssued, beforeChecks, beforePolls := f.issued, f.checks, f.polls
			f.mu.Unlock()
			check(t, step.ready, step.status)
			succeeds := step.ready && !step.caReject
			runError := api.run(t, workflowID, succeeds)
			if step.caReject && !strings.Contains(runError, "fixture CA rejected authorization") {
				t.Fatalf("workflow did not report the CA error: %s", runError)
			}
			f.mu.Lock()
			issued, checks, polls := f.issued, f.checks, f.polls
			remaining := slices.Clone(f.records[name])
			f.mu.Unlock()
			if succeeds {
				if issued != beforeIssued+1 || checks != beforeChecks+2 || polls != beforePolls+2 {
					t.Fatalf("issuance did not revalidate both domains: certificates=%d challenges=%d polls=%d", issued-beforeIssued, checks-beforeChecks, polls-beforePolls)
				}
				api.assertCertificate(t, domains, accountURL, int64(issued+99))
			} else if issued != beforeIssued {
				t.Fatalf("invalid TXT issued %d certificate(s)", issued-beforeIssued)
			}
			want := []string(nil)
			if step.value != "" {
				want = []string{step.value}
			}
			if !slices.Equal(remaining, want) {
				t.Fatalf("workflow changed the persistent TXT: %v", remaining)
			}
		})
	}
	t.Run("semantic-record-match", func(t *testing.T) {
		f.mu.Lock()
		f.records[name] = []string{"FIXTURE.EXAMPLE; POLICY=WILDCARD; accounturi=" + accountURL}
		f.mu.Unlock()
		check(t, true, "ready")
	})
	t.Run("records-removed-after-check", func(t *testing.T) {
		f.mu.Lock()
		f.records[name] = []string{value}
		f.mu.Unlock()
		check(t, true, "ready")
		f.mu.Lock()
		f.records[name] = nil
		beforeIssued := f.issued
		f.mu.Unlock()
		api.run(t, workflowID, false)
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.issued != beforeIssued || len(f.records[name]) != 0 {
			t.Fatal("issuance trusted stale readiness or restored the deleted TXT")
		}
	})
	t.Run("duplicate-account-parameter", func(t *testing.T) {
		f.mu.Lock()
		f.records[name] = []string{value + "; accounturi=" + accountURL}
		f.mu.Unlock()
		check(t, false, "mismatch")
	})
	otherCAID := api.create(t, "access", map[string]any{"name": "Other fixture CA", "provider": "acmeca", "reserve": "ca", "config": map[string]string{"endpoint": f.base + "/other-directory"}})
	for _, mutation := range []struct {
		name, key string
		value     any
	}{
		{name: "changed-email", key: "contactEmail", value: "changed@example.test"},
		{name: "missing-account", key: "dnsPersistAccountId", value: ""},
		{name: "changed-ca", key: "caProviderAccessId", value: otherCAID},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			changed := make(map[string]any, len(config))
			for k, v := range config {
				changed[k] = v
			}
			changed[mutation.key] = mutation.value
			f.mu.Lock()
			f.records[name] = []string{value}
			beforeAccounts, beforeIssued := f.accounts, f.issued
			f.mu.Unlock()
			id := api.create(t, "workflow", persistWorkflow(mutation.name, changed))
			api.run(t, id, false)
			f.mu.Lock()
			defer f.mu.Unlock()
			if f.accounts != beforeAccounts || f.issued != beforeIssued {
				t.Fatal("stale account binding registered an account or issued a certificate")
			}
		})
	}
	for _, shape := range []string{"unsupported-ca", "unknown-format", "custom-ca-cached-authorization"} {
		t.Run(shape, func(t *testing.T) {
			f.mu.Lock()
			f.unsupported = shape == "unsupported-ca"
			f.missingIssuer = shape == "unknown-format"
			f.cached = shape == "custom-ca-cached-authorization"
			f.mu.Unlock()
			var result persistPreparation
			if err := api.call("POST", "/api/acme-accounts/dns-persist/prepare", payload, &result); err == nil {
				t.Fatal("CA without usable dns-persist-01 challenge metadata was accepted")
			}
		})
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.accounts != 1 || f.queries == 0 || f.writes != 0 {
		t.Fatalf("unexpected lifecycle accounts=%d DNS queries=%d DNS mutations=%d", f.accounts, f.queries, f.writes)
	}
}

type persistAPI struct {
	base, token string
	client      *http.Client
}

func startPersistCore(t *testing.T, binary string, multiproc bool) *persistAPI {
	t.Helper()
	dataDir := t.TempDir()
	env := append(os.Environ(), "HTTP_PROXY=", "HTTPS_PROXY=", "ALL_PROXY=", "NO_PROXY=127.0.0.1,localhost", "CERTIMATE_PLUGIN_DIR="+t.TempDir(), "CERTIMATE_WORKFLOW_MULTIPROC="+strconv.FormatBool(multiproc))
	for _, args := range [][]string{{"migrate", "up", "--dir", dataDir}, {"superuser", "upsert", "persist@example.test", "Persist-fixture-password123!", "--dir", dataDir}} {
		cmd := exec.Command(binary, args...)
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("initialize core: %v %s", err, out)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	logFile, err := os.Create(filepath.Join(dataDir, "core.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "serve", "--http", address, "--dir", dataDir)
	cmd.Env, cmd.Stdout, cmd.Stderr = env, logFile, logFile
	if err := cmd.Start(); err != nil {
		logFile.Close()
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	t.Cleanup(func() {
		cmd.Process.Signal(os.Interrupt)
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			<-exited
		}
		logFile.Close()
		if t.Failed() {
			out, _ := os.ReadFile(logFile.Name())
			t.Log(string(out))
		}
	})
	api := &persistAPI{base: "http://" + address, client: &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{}}}
	deadline := time.Now().Add(time.Minute)
	for api.call("GET", "/api/health", nil, nil) != nil {
		if time.Now().After(deadline) {
			t.Fatal("core did not start")
		}
		time.Sleep(25 * time.Millisecond)
	}
	var auth struct {
		Token string `json:"token"`
	}
	if err := api.call("POST", "/api/collections/_superusers/auth-with-password", map[string]string{"identity": "persist@example.test", "password": "Persist-fixture-password123!"}, &auth); err != nil {
		t.Fatal(err)
	}
	api.token = auth.Token
	return api
}

func (a *persistAPI) call(method, path string, body, result any) error {
	var input io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, a.base+path, input)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", a.token)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, raw)
	}
	if strings.HasPrefix(path, "/api/acme-accounts/") || strings.HasPrefix(path, "/api/workflows/") {
		var envelope struct {
			Code int             `json:"code"`
			Msg  string          `json:"msg"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return err
		}
		if envelope.Code != 0 {
			return fmt.Errorf("API %d: %s", envelope.Code, envelope.Msg)
		}
		raw = envelope.Data
	}
	if result != nil {
		return json.Unmarshal(raw, result)
	}
	return nil
}

func (a *persistAPI) create(t *testing.T, collection string, body any) string {
	t.Helper()
	var result struct {
		ID string `json:"id"`
	}
	if err := a.call("POST", "/api/collections/"+collection+"/records", body, &result); err != nil {
		t.Fatal(err)
	}
	if result.ID == "" {
		t.Fatal("created record has no ID")
	}
	return result.ID
}

func persistWorkflow(name string, config map[string]any) map[string]any {
	node := func(id, kind string, cfg map[string]any) map[string]any {
		return map[string]any{"id": id, "type": kind, "data": map[string]any{"name": id, "config": cfg}}
	}
	graph := map[string]any{"nodes": []any{node("start", "start", map[string]any{}), node("apply", "bizApply", config), node("end", "end", map[string]any{})}}
	return map[string]any{"name": name, "trigger": "manual", "enabled": true, "hasContent": true, "hasDraft": false, "graphContent": graph, "graphDraft": graph}
}

func (a *persistAPI) run(t *testing.T, workflowID string, success bool) string {
	t.Helper()
	var run struct {
		ID string `json:"runId"`
	}
	if err := a.call("POST", "/api/workflows/"+workflowID+"/runs", map[string]string{"trigger": "manual"}, &run); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		var result struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if err := a.call("GET", "/api/collections/workflow_run/records/"+run.ID, nil, &result); err != nil {
			t.Fatal(err)
		}
		if result.Status == "succeeded" || result.Status == "failed" || result.Status == "canceled" {
			want := "failed"
			if success {
				want = "succeeded"
			}
			if result.Status != want {
				t.Fatalf("workflow %s, want %s: %s", result.Status, want, result.Error)
			}
			return result.Error
		}
		if time.Now().After(deadline) {
			t.Fatal("workflow did not terminate within 30 seconds")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func (a *persistAPI) assertCertificate(t *testing.T, domains []string, accountURL string, serial int64) {
	t.Helper()
	var response struct {
		Items []struct {
			Certificate string `json:"certificate"`
			PrivateKey  string `json:"privateKey"`
			AccountURL  string `json:"acmeAcctUrl"`
		} `json:"items"`
	}
	if err := a.call("GET", "/api/collections/certificate/records?perPage=1&sort=-created", nil, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 1 {
		t.Fatal("issued certificate was not stored")
	}
	stored := response.Items[0]
	if stored.AccountURL != accountURL {
		t.Fatalf("stored certificate uses account %q, want %q", stored.AccountURL, accountURL)
	}
	pair, err := tls.X509KeyPair([]byte(stored.Certificate), []byte(stored.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	if leaf.SerialNumber.Cmp(big.NewInt(serial)) != 0 {
		t.Fatalf("stored certificate serial=%s, want newly issued serial=%d", leaf.SerialNumber, serial)
	}
	got, want := slices.Clone(leaf.DNSNames), slices.Clone(domains)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("issued SANs %v, want %v", got, want)
	}
}

func (f *persistCAFixture) dns(w dns.ResponseWriter, r *dns.Msg) {
	f.mu.Lock()
	defer f.mu.Unlock()
	reply := new(dns.Msg)
	reply.SetReply(r)
	reply.Authoritative = true
	if r.Opcode != dns.OpcodeQuery {
		f.writes++
		reply.Rcode = dns.RcodeRefused
	} else {
		for _, q := range r.Question {
			if q.Qtype == dns.TypeTXT {
				f.queries++
				if f.dnsFailure {
					reply.Rcode = dns.RcodeServerFailure
					continue
				}
				for _, value := range f.records[q.Name] {
					reply.Answer = append(reply.Answer, &dns.TXT{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}, Txt: []string{value}})
				}
			}
		}
	}
	if err := w.WriteMsg(reply); err != nil {
		f.t.Error(err)
	}
}

func (f *persistCAFixture) acme(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Replay-Nonce", fmt.Sprintf("persist-fixture-%d", time.Now().UnixNano()))
	w.Header().Set("Content-Type", "application/json")
	reply := func(code int, value any) {
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(value); err != nil {
			f.t.Error(err)
		}
	}
	var payload map[string]any
	if r.Method == "POST" {
		var jws struct {
			Payload   string `json:"payload"`
			Protected string `json:"protected"`
		}
		if err := json.NewDecoder(r.Body).Decode(&jws); err != nil {
			reply(400, map[string]string{"detail": "invalid JWS"})
			return
		}
		protected, err := base64.RawURLEncoding.DecodeString(jws.Protected)
		var header struct {
			KID string `json:"kid"`
		}
		if err != nil || json.Unmarshal(protected, &header) != nil || (r.URL.Path != "/account" && header.KID != f.base+"/account/1") {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:accountDoesNotExist", "detail": "request uses the wrong ACME account"})
			return
		}
		raw, _ := base64.RawURLEncoding.DecodeString(jws.Payload)
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &payload); err != nil {
				reply(400, map[string]string{"detail": "invalid payload"})
				return
			}
		}
	}
	switch r.URL.Path {
	case "/directory":
		reply(200, map[string]any{"newNonce": f.base + "/nonce", "newAccount": f.base + "/account", "newOrder": f.base + "/new-order", "revokeCert": f.base + "/revoke", "meta": map[string]any{"termsOfService": f.base + "/tos", "caaIdentities": []string{"fixture.example"}, "profiles": map[string]string{"fixture": f.base + "/profile"}}})
		return
	case "/nonce":
		w.WriteHeader(200)
		return
	case "/account":
		f.accounts++
		w.Header().Set("Location", f.base+"/account/1")
		reply(201, map[string]any{"status": "valid", "contact": []string{"mailto:persist@example.test"}, "orders": f.base + "/orders"})
		return
	case "/account/1":
		reply(200, map[string]any{"status": "valid", "contact": []string{"mailto:persist@example.test"}})
		return
	case "/new-order":
		if payload["profile"] != "fixture" {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:malformed", "detail": "preparation and issuance must request the fixture profile"})
			return
		}
		identifiers, _ := payload["identifiers"].([]any)
		o := &persistOrder{}
		for _, raw := range identifiers {
			id, _ := raw.(map[string]any)
			value, _ := id["value"].(string)
			o.domains = append(o.domains, value)
			o.authorized = append(o.authorized, f.cached)
			o.submitted = append(o.submitted, false)
		}
		f.orders = append(f.orders, o)
		id := len(f.orders)
		w.Header().Set("Location", fmt.Sprintf("%s/order/%d", f.base, id))
		reply(201, f.orderResponse(id, o))
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		f.t.Errorf("unexpected ACME path %s", r.URL.Path)
		reply(404, map[string]string{"detail": "not found"})
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 || id > len(f.orders) {
		reply(404, map[string]string{"detail": "unknown order"})
		return
	}
	o := f.orders[id-1]
	switch parts[0] {
	case "order":
		reply(200, f.orderResponse(id, o))
	case "authz", "challenge":
		if len(parts) != 3 {
			reply(404, map[string]string{"detail": "unknown authorization"})
			return
		}
		i, err := strconv.Atoi(parts[2])
		if err != nil || i < 0 || i >= len(o.domains) {
			reply(404, map[string]string{"detail": "unknown authorization"})
			return
		}
		if parts[0] == "challenge" && payload != nil {
			f.checks++
			if f.reject {
				reply(400, map[string]string{"type": "urn:ietf:params:acme:error:unauthorized", "detail": "fixture CA rejected authorization"})
				return
			}
			name := "_validation-persist." + strings.TrimPrefix(o.domains[i], "*.") + "."
			value := "fixture.example; accounturi=" + f.base + "/account/1"
			matches := slices.Contains(f.records[name], value+"; policy=wildcard") || (!strings.HasPrefix(o.domains[i], "*.") && slices.Contains(f.records[name], value))
			if !matches {
				reply(400, map[string]string{"type": "urn:ietf:params:acme:error:unauthorized", "detail": "persistent TXT did not authorize this account"})
				return
			}
			o.submitted[i] = true
		} else if parts[0] == "challenge" && o.submitted[i] {
			f.polls++
			o.authorized[i] = true
		}
		state := "pending"
		if o.submitted[i] {
			state = "processing"
		}
		if o.authorized[i] {
			state = "valid"
		}
		challenge := map[string]any{"type": "dns-persist-01", "url": fmt.Sprintf("%s/challenge/%d/%d", f.base, id, i), "status": state, "issuer-domain-names": []string{"fixture.example"}}
		if f.missingIssuer {
			delete(challenge, "issuer-domain-names")
		}
		if !f.omitAccount {
			challenge["accounturi"] = f.base + "/account/1"
		}
		if f.unsupported {
			challenge["type"] = "dns-01"
			challenge["token"] = "fixture-token"
		}
		if parts[0] == "challenge" {
			w.Header().Set("Link", fmt.Sprintf("<%s/authz/%d/%d>;rel=\"up\"", f.base, id, i))
			reply(200, challenge)
		} else {
			challenges := []any{challenge}
			if f.cached {
				challenges = nil
			}
			reply(200, map[string]any{"status": state, "wildcard": strings.HasPrefix(o.domains[i], "*."), "identifier": map[string]string{"type": "dns", "value": strings.TrimPrefix(o.domains[i], "*.")}, "challenges": challenges})
		}
	case "finalize":
		if slices.Contains(o.authorized, false) {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:orderNotReady", "detail": "pending authorization"})
			return
		}
		encoded, _ := payload["csr"].(string)
		raw, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			reply(400, map[string]string{"detail": "invalid CSR encoding"})
			return
		}
		csr, err := x509.ParseCertificateRequest(raw)
		if err != nil || csr.CheckSignature() != nil {
			reply(400, map[string]string{"detail": "invalid CSR"})
			return
		}
		leaf := &x509.Certificate{SerialNumber: big.NewInt(int64(f.issued + 100)), Subject: csr.Subject, DNSNames: csr.DNSNames, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
		der, err := x509.CreateCertificate(rand.Reader, leaf, f.issuer, csr.PublicKey, f.key)
		if err != nil {
			f.t.Error(err)
			reply(500, map[string]string{"detail": "could not sign certificate"})
			return
		}
		o.chain = append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: f.issuer.Raw})...)
		f.issued++
		reply(200, f.orderResponse(id, o))
	case "certificate":
		w.Header().Set("Content-Type", "application/pem-certificate-chain")
		w.Write(o.chain)
	default:
		f.t.Errorf("unexpected ACME path %s", r.URL.Path)
		reply(404, map[string]string{"detail": "not found"})
	}
}

func (f *persistCAFixture) orderResponse(id int, o *persistOrder) map[string]any {
	state := "ready"
	if slices.Contains(o.authorized, false) {
		state = "pending"
	}
	if len(o.chain) > 0 {
		state = "valid"
	}
	identifiers, authorizations := make([]map[string]string, len(o.domains)), make([]string, len(o.domains))
	for i, domain := range o.domains {
		identifiers[i] = map[string]string{"type": "dns", "value": domain}
		authorizations[i] = fmt.Sprintf("%s/authz/%d/%d", f.base, id, i)
	}
	return map[string]any{"status": state, "identifiers": identifiers, "authorizations": authorizations, "finalize": fmt.Sprintf("%s/finalize/%d", f.base, id), "certificate": fmt.Sprintf("%s/certificate/%d", f.base, id)}
}
