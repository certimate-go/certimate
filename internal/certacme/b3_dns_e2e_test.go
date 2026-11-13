package certacme_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

type b3WildcardChallenge struct {
	token      string
	expected   string
	value      string
	authorized bool
}

type b3DNSFixture struct {
	t               *testing.T
	family          string
	proxyTargets    map[string]string
	proxyHosts      []string
	proxyTLSAddress string
	extraTrust      []byte
	credentialKey   *rsa.PrivateKey
	mu              sync.Mutex
	base            string
	domain          string
	token           string
	cname           string
	value           string
	expected        string
	authorized      bool
	wildcard        *b3WildcardChallenge
	mode            string
	creates         int
	deletes         int
	issued          int
	denied          int
	txtResponses    int
	chain           []byte
	issuer          *x509.Certificate
	key             *ecdsa.PrivateKey
}

func TestB3DNSWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_B3_DNS_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_B3_DNS_CORE to the core executable")
	}
	if runtime.GOOS != "linux" {
		dir := t.TempDir()
		cmd := exec.Command("go", "test", "-c", "-o", filepath.Join(dir, "dns.test"), ".")
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm64", "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build DNS test: %v %s", err, out)
		}
		data, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "core"), data, 0o755); err != nil {
			t.Fatal(err)
		}
		cmd = exec.Command("docker", "run", "--rm", "-v", dir+":/fixture:ro", "-e", "CERTIMATE_B3_DNS_CORE=/fixture/core", "-e", "CERTIMATE_B3_DNS_FAMILIES="+os.Getenv("CERTIMATE_B3_DNS_FAMILIES"), "alpine:latest", "/fixture/dns.test", "-test.run=^TestB3DNSWorkflow$", "-test.v", "-test.timeout=15m")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	families := []string{"linode", "qingcloud", "bunny", "cloudflare", "cpanel", "digitalocean", "gcore", "googlecloud", "netlify", "oraclecloud", "vercel", "yandexcloud"}
	if filter := os.Getenv("CERTIMATE_B3_DNS_FAMILIES"); filter != "" {
		families = strings.Split(filter, ",")
	}
	for _, family := range families {
		t.Run(family, func(t *testing.T) { runDNSWorkflow(t, binary, family, dnsWorkflowPlugins{}) })
	}
}

type dnsWorkflowPlugins struct {
	dir              string
	accessByProvider map[string]string
	configureDNS     func(*testing.T, *b3DNSFixture) b3DNSProviderConfig
	wildcardOrder    bool
	cname            bool
}

func runDNSWorkflow(t *testing.T, binary, family string, plugins dnsWorkflowPlugins) {
	f := &b3DNSFixture{t: t, family: family, proxyTargets: map[string]string{}, domain: "b3-" + family + ".example.test", token: "b3-fixture-challenge-token"}
	domains := []string{f.domain}
	if plugins.wildcardOrder {
		f.domain = "example.com"
		f.wildcard = &b3WildcardChallenge{token: "b3-fixture-wildcard-token"}
		domains = []string{f.domain, "*." + f.domain}
	}
	if plugins.cname {
		f.cname = "_acme-challenge.delegated.example.test."
	}
	var err error
	f.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "B3 fixture issuer"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, issuer, issuer, &f.key.PublicKey, f.key)
	if err != nil {
		t.Fatal(err)
	}
	f.issuer, err = x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	ca := httptest.NewTLSServer(http.HandlerFunc(f.acme))
	defer ca.Close()
	f.base = ca.URL
	udp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dnsServer := &dns.Server{PacketConn: udp, Handler: dns.HandlerFunc(f.dns)}
	go dnsServer.ActivateAndServe()
	defer dnsServer.Shutdown()
	var providerConfig b3DNSProviderConfig
	ok := plugins.configureDNS != nil
	if ok {
		providerConfig = plugins.configureDNS(t, f)
	} else {
		providerConfig, ok = b4DNSConfig(t, f)
	}
	if !ok {
		providerConfig, ok = b5DNSConfig(f)
	}
	if !ok {
		providerConfig, ok = b6B8DNSConfig(t, f)
	}
	if !ok {
		providerConfig = b3DNSConfig(t, f, family)
	}
	proxyURL, caPath := f.proxy(t)
	trusted, err := os.ReadFile(caPath)
	if err != nil {
		t.Fatal(err)
	}
	trusted = append(trusted, f.extraTrust...)
	trusted = append(trusted, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate().Raw})...)
	if err := os.WriteFile(caPath, trusted, 0o600); err != nil {
		t.Fatal(err)
	}
	pluginDir := plugins.dir
	if pluginDir == "" {
		pluginDir = t.TempDir()
	}
	env := append(os.Environ(), "HTTPS_PROXY="+proxyURL, "HTTP_PROXY="+proxyURL, "NO_PROXY=127.0.0.1,localhost", "SSL_CERT_FILE="+caPath, "LEGO_CA_CERTIFICATES="+caPath, "CERTIMATE_PLUGIN_DIR="+pluginDir)
	env = append(env, providerConfig.Env...)
	dataDir := t.TempDir()
	log := new(bytes.Buffer)
	for _, args := range [][]string{{"migrate", "up", "--dir", dataDir}, {"superuser", "upsert", "dns@example.test", "DNS-fixture-password123!", "--dir", dataDir}} {
		cmd := exec.Command(binary, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup: %v %s", err, out)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	cmd := exec.Command(binary, "serve", "--http", address, "--dir", dataDir)
	cmd.Env = env
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			<-exited
		}
		if t.Failed() {
			t.Log(log.String())
		}
	}()
	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{}}
	base := "http://" + address
	token := ""
	call := func(method, path string, body any, result any) error {
		var payload io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			payload = bytes.NewReader(b)
		}
		req, _ := http.NewRequest(method, base+path, payload)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d %s", resp.StatusCode, b)
		}
		if result != nil {
			return json.NewDecoder(resp.Body).Decode(result)
		}
		return nil
	}
	deadline := time.Now().Add(2 * time.Minute)
	for call("GET", "/api/health", nil, nil) != nil {
		if time.Now().After(deadline) {
			t.Fatal("core not ready")
		}
		time.Sleep(25 * time.Millisecond)
	}
	var auth struct {
		Token string `json:"token"`
	}
	if err := call("POST", "/api/collections/_superusers/auth-with-password", map[string]string{"identity": "dns@example.test", "password": "DNS-fixture-password123!"}, &auth); err != nil {
		t.Fatal(err)
	}
	token = auth.Token
	if len(plugins.accessByProvider) != 0 {
		var catalog struct {
			Data []struct {
				Source             string `json:"source"`
				ProviderType       string `json:"providerType"`
				AccessProviderType string `json:"accessProviderType"`
			} `json:"data"`
		}
		if err := call("GET", "/api/plugin-catalog", nil, &catalog); err != nil {
			t.Fatal(err)
		}
		providers := make(map[string]string, len(catalog.Data))
		for _, entry := range catalog.Data {
			if entry.Source == "plugin" {
				providers[entry.ProviderType] = entry.AccessProviderType
			}
		}
		for provider, accessType := range plugins.accessByProvider {
			if providers[provider] != accessType {
				t.Fatalf("expected installed plugin %q to use access %q, got %q", provider, accessType, providers[provider])
			}
		}
		t.Logf("%d plugins loaded with the expected access types", len(plugins.accessByProvider))
	}
	create := func(collection string, body any) string {
		var result struct {
			ID string `json:"id"`
		}
		if err := call("POST", "/api/collections/"+collection+"/records", body, &result); err != nil {
			t.Fatal(err)
		}
		return result.ID
	}
	accessConfig, provider := providerConfig.Access, providerConfig.Provider
	accessID := create("access", map[string]any{"name": "B3 " + family, "provider": family, "config": accessConfig})
	caID := create("access", map[string]any{"name": "B3 ACME", "provider": "acmeca", "config": map[string]any{"endpoint": ca.URL + "/directory"}})
	steps := []string{"initial", "auth-failure", "recovery"}
	if f.wildcard != nil {
		steps = []string{"initial", "auth-failure", "partial-present-failure", "recovery"}
	}
	for _, step := range steps {
		t.Run(step, func(t *testing.T) {
			f.mu.Lock()
			f.mode = step
			f.authorized = false
			if f.wildcard != nil {
				f.wildcard.authorized = false
			}
			f.chain = nil
			f.txtResponses = 0
			creates, deletes, issued, denied := f.creates, f.deletes, f.issued, f.denied
			f.mu.Unlock()
			node := func(id, kind string, config map[string]any) map[string]any {
				return map[string]any{"id": id, "type": kind, "data": map[string]any{"name": id, "config": config}}
			}
			config := map[string]any{"domains": strings.Join(domains, ";"), "contactEmail": "dns@example.test", "challengeType": "dns-01", "provider": provider, "providerAccessId": accessID, "caProvider": "acmeca", "caProviderAccessId": caID, "keyAlgorithm": "EC256", "nameservers": udp.LocalAddr().String(), "dnsPropagationTimeout": 10, "dnsTTL": 60, "disableARI": true, "disableFollowCNAME": true}
			for key, value := range providerConfig.Apply {
				config[key] = value
			}
			graph := map[string]any{"nodes": []any{node("start", "start", map[string]any{}), node("apply", "bizApply", config), node("end", "end", map[string]any{})}}
			workflowID := create("workflow", map[string]any{"name": "B3 DNS " + family + " " + step, "trigger": "manual", "enabled": true, "hasContent": true, "hasDraft": false, "graphContent": graph, "graphDraft": graph})
			var run struct {
				Data struct {
					RunID string `json:"runId"`
				} `json:"data"`
			}
			if err := call("POST", "/api/workflows/"+workflowID+"/runs", map[string]string{"trigger": "manual"}, &run); err != nil {
				t.Fatal(err)
			}
			var result struct {
				Status string `json:"status"`
				Error  string `json:"error"`
			}
			deadline := time.Now().Add(45 * time.Second)
			for {
				if err := call("GET", "/api/collections/workflow_run/records/"+run.Data.RunID, nil, &result); err != nil {
					t.Fatal(err)
				}
				if result.Status == "succeeded" || result.Status == "failed" || result.Status == "canceled" {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("workflow timeout")
				}
				time.Sleep(30 * time.Millisecond)
			}
			if f.wildcard != nil && (result.Status == "succeeded" || step == "partial-present-failure") {
				query := new(dns.Msg).SetQuestion(f.challengeRecordName(), dns.TypeTXT)
				client := &dns.Client{Timeout: time.Second}
				reply, _, err := client.Exchange(query, udp.LocalAddr().String())
				if err != nil {
					t.Fatalf("query TXT after issuance: %v", err)
				}
				if reply.Rcode != dns.RcodeSuccess || len(reply.Answer) != 0 {
					t.Errorf("DNS still returns challenge TXT after the workflow: %v", reply)
				}
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			if step == "auth-failure" {
				if result.Status != "failed" || f.creates != creates || f.issued != issued || f.denied <= denied {
					t.Fatalf("authorization failure status=%s creates=%d issued=%d error=%s", result.Status, f.creates, f.issued, result.Error)
				}
				t.Log("authorization failure rejected without issuance")
				return
			}
			var remaining []string
			if f.value != "" {
				remaining = append(remaining, f.value)
			}
			if f.wildcard != nil && f.wildcard.value != "" {
				remaining = append(remaining, f.wildcard.value)
			}
			if step == "partial-present-failure" {
				if result.Status != "failed" || f.denied-denied != 1 || f.creates-creates != 1 || f.deletes-deletes != 1 || f.issued != issued || len(remaining) != 0 {
					t.Fatalf("partial failure status=%s creates=%d deletes=%d issued=%d denied=%d residual=%q error=%s", result.Status, f.creates-creates, f.deletes-deletes, f.issued-issued, f.denied-denied, remaining, result.Error)
				}
				t.Log("one TXT write was rejected and the other TXT was cleaned up without issuance")
				return
			}
			if result.Status != "succeeded" {
				t.Fatalf("workflow %s: %s", result.Status, result.Error)
			}
			if f.creates-creates != len(domains) || f.deletes-deletes != len(domains) || f.issued-issued != 1 || len(remaining) != 0 {
				t.Fatalf("DNS lifecycle creates=%d deletes=%d issued=%d residual=%q; want %d creates and deletes, one certificate and no residual TXT values", f.creates-creates, f.deletes-deletes, f.issued-issued, remaining, len(domains))
			}
			if f.txtResponses == 0 {
				t.Fatal("no DNS query received the expected TXT value")
			}
			var certificates struct {
				Items []struct {
					Certificate   string `json:"certificate"`
					PrivateKey    string `json:"privateKey"`
					WorkflowRunID string `json:"workflowRunRef"`
				} `json:"items"`
			}
			if err := call("GET", "/api/collections/certificate/records?perPage=100", nil, &certificates); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, record := range certificates.Items {
				if strings.TrimSpace(record.Certificate) == strings.TrimSpace(string(f.chain)) {
					leaf, _ := pem.Decode([]byte(record.Certificate))
					if leaf == nil {
						t.Fatal("persisted certificate not PEM")
					}
					parsed, err := x509.ParseCertificate(leaf.Bytes)
					if err != nil || parsed.VerifyHostname(f.domain) != nil {
						t.Fatal("persisted certificate SAN mismatch")
					}
					if f.wildcard != nil && (len(parsed.DNSNames) != len(domains) || !slices.Contains(parsed.DNSNames, domains[1])) {
						t.Fatalf("persisted certificate SANs = %v, want %v", parsed.DNSNames, domains)
					}
					if _, err := tls.X509KeyPair([]byte(record.Certificate), []byte(record.PrivateKey)); err != nil {
						t.Fatal(err)
					}
					found = true
				}
			}
			if !found {
				t.Fatal("issued chain and matching key not persisted")
			}
			t.Logf("workflow %s issued and persisted certificate for %v; %d TXT values created and deleted, with %d DNS responses", run.Data.RunID, domains, len(domains), f.txtResponses)
		})
	}
}

func (f *b3DNSFixture) challengeRecordName() string {
	if f.cname != "" {
		return f.cname
	}
	return "_acme-challenge." + f.domain + "."
}

func (f *b3DNSFixture) dns(w dns.ResponseWriter, r *dns.Msg) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	matchedTXT := false
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeSOA:
			if q.Name == "example.test." || (f.wildcard != nil && q.Name == "example.com.") {
				m.Answer = append(m.Answer, &dns.SOA{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60}, Ns: "ns.example.test.", Mbox: "admin.example.test.", Serial: 1, Refresh: 60, Retry: 60, Expire: 3600, Minttl: 60})
			}
		case dns.TypeCNAME:
			if f.cname != "" && q.Name == "_acme-challenge."+f.domain+"." {
				m.Answer = append(m.Answer, &dns.CNAME{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 60}, Target: f.cname})
			}
		case dns.TypeTXT:
			if f.wildcard != nil && q.Name == f.challengeRecordName() && f.wildcard.value != "" {
				m.Answer = append(m.Answer, &dns.TXT{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}, Txt: []string{f.wildcard.value}})
				matchedTXT = f.wildcard.value == f.wildcard.expected
			}
			if q.Name == f.challengeRecordName() && f.value != "" {
				m.Answer = append(m.Answer, &dns.TXT{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}, Txt: []string{f.value}})
				matchedTXT = matchedTXT || f.value == f.expected
			}
		case dns.TypeNS:
			m.Answer = append(m.Answer, &dns.NS{Hdr: dns.RR_Header{Name: "example.test.", Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 60}, Ns: "ns.example.test."})
		case dns.TypeA:
			m.Answer = append(m.Answer, &dns.A{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}, A: net.ParseIP("127.0.0.1")})
		}
	}
	if err := w.WriteMsg(m); err != nil {
		f.t.Error(err)
	} else if matchedTXT {
		f.txtResponses++
	}
}

func (f *b3DNSFixture) acme(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Replay-Nonce", fmt.Sprintf("fixture-nonce-%d", time.Now().UnixNano()))
	w.Header().Set("Content-Type", "application/json")
	reply := func(status int, v any) { w.WriteHeader(status); json.NewEncoder(w).Encode(v) }
	var jws struct {
		Payload   string `json:"payload"`
		Protected string `json:"protected"`
	}
	var payload map[string]any
	if r.Method == "POST" {
		json.NewDecoder(r.Body).Decode(&jws)
		raw, _ := base64.RawURLEncoding.DecodeString(jws.Payload)
		json.Unmarshal(raw, &payload)
	}
	authorized := f.authorized && (f.wildcard == nil || f.wildcard.authorized)
	order := func(status string) map[string]any {
		identifiers := []map[string]string{{"type": "dns", "value": f.domain}}
		authorizations := []string{f.base + "/authz"}
		if f.wildcard != nil {
			identifiers = append(identifiers, map[string]string{"type": "dns", "value": "*." + f.domain})
			authorizations = append(authorizations, f.base+"/authz-wildcard")
		}
		v := map[string]any{"status": status, "identifiers": identifiers, "authorizations": authorizations, "finalize": f.base + "/finalize"}
		if status == "valid" {
			v["certificate"] = f.base + "/certificate"
		}
		return v
	}
	switch r.URL.Path {
	case "/directory":
		reply(200, map[string]any{"newNonce": f.base + "/nonce", "newAccount": f.base + "/account", "newOrder": f.base + "/new-order", "revokeCert": f.base + "/revoke", "keyChange": f.base + "/key-change", "meta": map[string]string{"termsOfService": f.base + "/tos"}})
	case "/nonce":
		w.WriteHeader(200)
	case "/account":
		raw, _ := base64.RawURLEncoding.DecodeString(jws.Protected)
		var p struct {
			JWK map[string]any `json:"jwk"`
		}
		json.Unmarshal(raw, &p)
		if p.JWK != nil {
			canonical, _ := json.Marshal(p.JWK)
			thumb := sha256.Sum256(canonical)
			digest := sha256.Sum256([]byte(f.token + "." + base64.RawURLEncoding.EncodeToString(thumb[:])))
			f.expected = base64.RawURLEncoding.EncodeToString(digest[:])
			if f.wildcard != nil {
				digest := sha256.Sum256([]byte(f.wildcard.token + "." + base64.RawURLEncoding.EncodeToString(thumb[:])))
				f.wildcard.expected = base64.RawURLEncoding.EncodeToString(digest[:])
			}
		}
		w.Header().Set("Location", f.base+"/account/1")
		reply(201, map[string]any{"status": "valid", "contact": []string{"mailto:dns@example.test"}, "orders": f.base + "/orders"})
	case "/account/1":
		reply(200, map[string]any{"status": "valid", "contact": []string{"mailto:dns@example.test"}})
	case "/new-order":
		if f.wildcard != nil {
			identifiers, _ := payload["identifiers"].([]any)
			want := []string{f.domain, "*." + f.domain}
			var got []string
			for _, item := range identifiers {
				id, _ := item.(map[string]any)
				value, _ := id["value"].(string)
				got = append(got, value)
			}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				f.t.Errorf("ACME order identifiers = %v, want %v", got, want)
			}
		}
		w.Header().Set("Location", f.base+"/order")
		reply(201, order("pending"))
	case "/order":
		state := "pending"
		if authorized {
			state = "ready"
		}
		if len(f.chain) > 0 {
			state = "valid"
		}
		reply(200, order(state))
	case "/authz":
		state := "pending"
		if f.authorized {
			state = "valid"
		}
		reply(200, map[string]any{"status": state, "identifier": map[string]string{"type": "dns", "value": f.domain}, "challenges": []any{map[string]string{"type": "dns-01", "url": f.base + "/challenge", "token": f.token, "status": state}}})
	case "/challenge":
		if f.value == "" || f.value != f.expected {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:unauthorized", "detail": "TXT does not match account key authorization"})
			return
		}
		f.authorized = true
		reply(200, map[string]string{"type": "dns-01", "url": f.base + "/challenge", "token": f.token, "status": "valid"})
	case "/authz-wildcard":
		state := "pending"
		if f.wildcard.authorized {
			state = "valid"
		}
		reply(200, map[string]any{"status": state, "wildcard": true, "identifier": map[string]string{"type": "dns", "value": f.domain}, "challenges": []any{map[string]string{"type": "dns-01", "url": f.base + "/challenge-wildcard", "token": f.wildcard.token, "status": state}}})
	case "/challenge-wildcard":
		if f.wildcard.value == "" || f.wildcard.value != f.wildcard.expected {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:unauthorized", "detail": "wildcard TXT does not match account key authorization"})
			return
		}
		f.wildcard.authorized = true
		reply(200, map[string]string{"type": "dns-01", "url": f.base + "/challenge-wildcard", "token": f.wildcard.token, "status": "valid"})
	case "/finalize":
		if !authorized {
			reply(400, map[string]string{"type": "urn:ietf:params:acme:error:orderNotReady"})
			return
		}
		encoded, _ := payload["csr"].(string)
		raw, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			f.t.Error(err)
			w.WriteHeader(400)
			return
		}
		csr, err := x509.ParseCertificateRequest(raw)
		if err != nil || csr.CheckSignature() != nil {
			f.t.Error("invalid CSR")
			w.WriteHeader(400)
			return
		}
		leaf := &x509.Certificate{SerialNumber: big.NewInt(int64(f.issued + 10)), Subject: csr.Subject, DNSNames: csr.DNSNames, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
		der, err := x509.CreateCertificate(rand.Reader, leaf, f.issuer, csr.PublicKey, f.key)
		if err != nil {
			f.t.Error(err)
			w.WriteHeader(500)
			return
		}
		f.chain = append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: f.issuer.Raw})...)
		f.issued++
		reply(200, order("valid"))
	case "/certificate":
		w.Header().Set("Content-Type", "application/pem-certificate-chain")
		w.Write(f.chain)
	default:
		f.t.Errorf("unhandled ACME %s %s", r.Method, r.URL.Path)
		w.WriteHeader(404)
	}
}

func (f *b3DNSFixture) cloud(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	reply := func(v any) { json.NewEncoder(w).Encode(v) }
	if f.mode == "auth-failure" {
		f.denied++
		w.WriteHeader(403)
		reply(map[string]string{"message": "fixture forbidden"})
		return
	}
	if f.b4DNSCloud(w, r) || f.b5DNSCloud(w, r) || f.b6B8DNSCloud(w, r) || f.remainingCloud(w, r) {
		return
	}
	if strings.Contains(r.Host, "linode") {
		if r.Header.Get("Authorization") != "Bearer fixture-dns-token" {
			f.t.Error("Linode credential mismatch")
			w.WriteHeader(403)
			return
		}
		switch {
		case r.Method == "GET" && r.URL.Path == "/v4/domains":
			reply(map[string]any{"data": []any{map[string]any{"id": 42, "domain": "example.test"}}, "page": 1, "pages": 1, "results": 1})
		case r.Method == "POST" && r.URL.Path == "/v4/domains/42/records":
			var body struct{ Name, Type, Target string }
			json.NewDecoder(r.Body).Decode(&body)
			if body.Name != "_acme-challenge."+f.domain || body.Type != "TXT" {
				f.t.Error("Linode wrong TXT target")
			}
			f.value = body.Target
			f.creates++
			reply(map[string]any{"id": 7})
		case r.Method == "DELETE" && r.URL.Path == "/v4/domains/42/records/7":
			f.value = ""
			f.deletes++
			reply(map[string]any{})
		default:
			f.t.Errorf("unhandled Linode %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
		return
	}
	if !strings.Contains(r.Host, "routewize") {
		f.t.Errorf("unexpected cloud host %s", r.Host)
		w.WriteHeader(404)
		return
	}
	mac := hmac.New(sha256.New, []byte("fixture-dns-secret"))
	mac.Write([]byte(r.Method + "\n" + r.Header.Get("Date") + "\n" + r.URL.RequestURI()))
	want := "QC-HMAC-SHA256 fixture-dns-id:" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if r.Header.Get("Authorization") != want {
		f.t.Error("QingCloud signature mismatch")
		w.WriteHeader(403)
		return
	}
	switch r.URL.Path {
	case "/v1/record/":
		var body struct {
			DomainName string `json:"domain_name"`
			ZoneName   string `json:"zone_name"`
			Type       string `json:"type"`
			Record     []struct {
				Values []struct {
					Value string `json:"value"`
				} `json:"values"`
			} `json:"record"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.DomainName != "_acme-challenge."+f.domain+"." || body.ZoneName != "example.test." || body.Type != "TXT" || len(body.Record) != 1 || len(body.Record[0].Values) != 1 {
			f.t.Error("QingCloud wrong TXT target")
			w.WriteHeader(400)
			return
		}
		f.value = body.Record[0].Values[0].Value
		f.creates++
		reply(map[string]any{"code": 0, "domain_record_id": 7})
	case "/v1/change_record_status/":
		var body struct {
			IDs            []int `json:"ids"`
			Action, Target string
		}
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.IDs) != 1 || body.IDs[0] != 7 || body.Action != "delete" || body.Target != "record" {
			f.t.Error("QingCloud wrong deletion")
		}
		f.value = ""
		f.deletes++
		reply(map[string]any{"code": 0})
	default:
		f.t.Errorf("unhandled QingCloud %s", r.URL.Path)
		w.WriteHeader(404)
	}
}

func (f *b3DNSFixture) proxy(t *testing.T) (string, string) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(99), Subject: pkix.Name{CommonName: "B3 DNS proxy"}, DNSNames: append([]string{"api.linode.com", "api.bunny.net", "api.cloudflare.com", "cpanel.example.test", "api.digitalocean.com", "api.gcore.com", "dns.googleapis.com", "www.googleapis.com", "oauth2.googleapis.com", "api.netlify.com", "dns.us-phoenix-1.oraclecloud.com", "dns.us-phoenix-1.oci.oraclecloud.com", "api.vercel.com"}, f.proxyHosts...), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	pair, err := tls.X509KeyPair(certPEM, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	if err != nil {
		t.Fatal(err)
	}
	target := httptest.NewUnstartedServer(http.HandlerFunc(f.cloud))
	if f.proxyTLSAddress != "" {
		_ = target.Listener.Close()
		target.Listener, err = net.Listen("tcp", f.proxyTLSAddress)
		if err != nil {
			t.Fatal(err)
		}
	}
	target.TLS = &tls.Config{Certificates: []tls.Certificate{pair}}
	target.StartTLS()
	t.Cleanup(target.Close)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "CONNECT" {
			f.cloud(w, r)
			return
		}
		host, _, _ := net.SplitHostPort(r.Host)
		allowed := false
		for _, name := range cert.DNSNames {
			if host == name {
				allowed = true
			}
		}
		if !allowed {
			t.Errorf("unexpected CONNECT %s", r.Host)
			w.WriteHeader(502)
			return
		}
		targetAddress := target.Listener.Addr().String()
		if override := f.proxyTargets[r.Host]; override != "" {
			targetAddress = override
		}
		conn, err := net.Dial("tcp", targetAddress)
		if err != nil {
			t.Error(err)
			w.WriteHeader(502)
			return
		}
		hijacker := w.(http.Hijacker)
		client, buf, err := hijacker.Hijack()
		if err != nil {
			conn.Close()
			t.Error(err)
			return
		}
		buf.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		buf.Flush()
		go func() { defer conn.Close(); defer client.Close(); io.Copy(conn, client) }()
		go func() { defer conn.Close(); defer client.Close(); io.Copy(client, conn) }()
	}))
	t.Cleanup(proxy.Close)
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return proxy.URL, path
}
