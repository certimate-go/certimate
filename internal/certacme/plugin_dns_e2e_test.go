package certacme_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func urlQueryEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\'' || c == '=' || c == '&' || c == ' ' || c == '<' || c == '>' {
			b.WriteString(fmt.Sprintf("%%%02X", c))
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

type pluginAPIFixture struct {
	t            *testing.T
	dns          *b3DNSFixture
	mu           sync.Mutex
	presents     int
	cleanups     int
	deploys      int
	notifies     int
	deployedCert string
	subject      string
	message      string
}

func (a *pluginAPIFixture) handle(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer plugin-dns-secret" && r.URL.Path != "/notify" {
		a.t.Errorf("plugin API %s: missing bearer token", r.URL.Path)
		w.WriteHeader(403)
		return
	}
	var body struct {
		Action      string `json:"action"`
		Fqdn        string `json:"fqdn"`
		Token       string `json:"token"`
		KeyAuth     string `json:"keyAuth"`
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"privateKey"`
		Subject     string `json:"subject"`
		Message     string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	switch r.URL.Path {
	case "/present":
		if !strings.Contains(body.Fqdn, "mockdns.example.test") {
			a.t.Errorf("present fqdn=%q", body.Fqdn)
		}
		digest := sha256.Sum256([]byte(body.KeyAuth))
		a.presents++
		a.mu.Unlock()
		a.dns.mu.Lock()
		a.dns.value = base64.RawURLEncoding.EncodeToString(digest[:])
		a.dns.mu.Unlock()
		a.mu.Lock()
	case "/cleanup":
		if !strings.Contains(body.Fqdn, "mockdns.example.test") {
			a.t.Errorf("cleanup fqdn=%q", body.Fqdn)
		}
		a.cleanups++
		a.mu.Unlock()
		a.dns.mu.Lock()
		a.dns.value = ""
		a.dns.mu.Unlock()
		a.mu.Lock()
	case "/deploy":
		a.deploys++
		a.deployedCert = body.Certificate
	case "/notify":
		a.notifies++
		a.subject = body.Subject
		a.message = body.Message
	default:
		a.t.Errorf("unhandled plugin API %s %s", r.Method, r.URL.Path)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("{}"))
}

func TestPluginDNSWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_PLUGIN_E2E_CORE")
	pluginDir := ""
	if binary == "" {
		if runtime.GOOS == "windows" {
			t.Skip("plugin fixture workflow relies on unix process signals")
		}
		_, thisFile, _, _ := runtime.Caller(0)
		coreRoot, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", ".."))
		if err != nil {
			t.Fatal(err)
		}
		workspace := filepath.Dir(coreRoot)
		buildDir := t.TempDir()
		cross := append(os.Environ(), "CGO_ENABLED=0")
		if runtime.GOOS != "linux" {
			cross = append(cross, "GOOS=linux", "GOARCH=arm64")
		}
		build := func(name, sourceDir string) string {
			out := filepath.Join(buildDir, name)
			cmd := exec.Command("go", "build", "-o", out, ".")
			cmd.Dir = sourceDir
			cmd.Env = cross
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build %s: %v %s", name, err, output)
			}
			return out
		}
		build("core", coreRoot)
		for _, name := range []string{"mockdns", "mockdns-deploy", "mocknotify"} {
			build(name, filepath.Join(workspace, "plugins", name))
		}
		if runtime.GOOS != "linux" {
			dir := t.TempDir()
			cmd := exec.Command("go", "test", "-c", "-o", filepath.Join(dir, "plugin.test"), ".")
			cmd.Env = cross
			cmd.Dir = filepath.Join(coreRoot, "internal", "certacme")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build plugin e2e test: %v %s", err, out)
			}
			for _, name := range []string{"core", "mockdns", "mockdns-deploy", "mocknotify"} {
				data, err := os.ReadFile(filepath.Join(buildDir, name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name), data, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"mockdns", "mockdns-deploy", "mocknotify"} {
				source, err := os.ReadFile(filepath.Join(workspace, "plugins", name, "manifest.json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name+".manifest.json"), source, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			wrap := filepath.Join(dir, "wrap")
			script := "#!/bin/sh\nmkdir -p /work/plugins\nfor n in mockdns mockdns-deploy mocknotify; do\n  mkdir -p /work/plugins/$n\n  cp /fixture/$n /work/plugins/$n/$n\n  cp /fixture/$n.manifest.json /work/plugins/$n/manifest.json\ndone\nexec /fixture/plugin.test -test.run='^TestPluginDNSWorkflow$' -test.v -test.timeout=15m\n"
			if err := os.WriteFile(wrap, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			cmd = exec.Command("docker", "run", "--rm", "-v", dir+":/fixture:ro", "-e", "CERTIMATE_PLUGIN_E2E_CORE=/fixture/core", "-e", "CERTIMATE_PLUGIN_E2E_PLUGINS=/work/plugins", "alpine:latest", "/bin/sh", "/fixture/wrap")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		binary = filepath.Join(buildDir, "core")
		pluginStage := filepath.Join(buildDir, "plugins")
		for _, name := range []string{"mockdns", "mockdns-deploy", "mocknotify"} {
			dir := filepath.Join(pluginStage, name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(buildDir, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o755); err != nil {
				t.Fatal(err)
			}
			manifest, err := os.ReadFile(filepath.Join(workspace, "plugins", name, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifest, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		pluginDir = pluginStage
	}

	if pluginDir == "" {
		pluginDir = os.Getenv("CERTIMATE_PLUGIN_E2E_PLUGINS")
	}

	f := &b3DNSFixture{t: t, family: "mockdns", proxyTargets: map[string]string{}, domain: "mockdns.example.test", token: "plugin-fixture-challenge-token"}
	var err error
	f.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Plugin fixture issuer"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
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

	api := &pluginAPIFixture{t: t, dns: f}
	apiServer := httptest.NewServer(http.HandlerFunc(api.handle))
	defer apiServer.Close()

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate().Raw})
	if err := os.WriteFile(caPath, caPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	env := []string{}
	for _, entry := range os.Environ() {
		switch {
		case strings.HasPrefix(entry, "HTTPS_PROXY="), strings.HasPrefix(entry, "HTTP_PROXY="), strings.HasPrefix(entry, "https_proxy="), strings.HasPrefix(entry, "http_proxy="):
		default:
			env = append(env, entry)
		}
	}
	env = append(env, "NO_PROXY=127.0.0.1,localhost", "SSL_CERT_FILE="+caPath, "CERTIMATE_PLUGIN_DIR="+pluginDir)

	dataDir := t.TempDir()
	log := new(bytes.Buffer)
	for _, args := range [][]string{{"migrate", "up", "--dir", dataDir}, {"superuser", "upsert", "plugin@example.test", "Plugin-fixture-pass123!", "--dir", dataDir}} {
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
	serve := exec.Command(binary, "serve", "--http", address, "--dir", dataDir)
	serve.Env = env
	serve.Stdout = log
	serve.Stderr = log
	if err := serve.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- serve.Wait() }()
	defer func() {
		serve.Process.Signal(os.Interrupt)
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			serve.Process.Kill()
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
	if err := call("POST", "/api/collections/_superusers/auth-with-password", map[string]string{"identity": "plugin@example.test", "password": "Plugin-fixture-pass123!"}, &auth); err != nil {
		t.Fatal(err)
	}
	token = auth.Token

	t.Run("catalog-and-schemas", func(t *testing.T) {
		var catalog struct {
			Data []struct {
				Source       string `json:"source"`
				ProviderType string `json:"providerType"`
				Capability   string `json:"capability"`
				OwnsAccess   bool   `json:"ownsAccess"`
			} `json:"data"`
		}
		if err := call("GET", "/api/plugin-catalog", nil, &catalog); err != nil {
			t.Fatal(err)
		}
		entries := map[string]struct {
			capability string
			ownsAccess bool
		}{}
		for _, entry := range catalog.Data {
			if entry.Source == "plugin" {
				entries[entry.ProviderType] = struct {
					capability string
					ownsAccess bool
				}{entry.Capability, entry.OwnsAccess}
			}
		}
		if got := entries["mockdns"]; got.capability != "dns" || !got.ownsAccess {
			t.Fatalf("expected mockdns dns owner, got %+v (all=%+v)", got, entries)
		}
		if got := entries["mockdns-deploy"]; got.capability != "deploy" || got.ownsAccess {
			t.Fatalf("expected mockdns-deploy deploy reuser, got %+v", got)
		}
		if got := entries["mocknotify"]; got.capability != "notify" || got.ownsAccess {
			t.Fatalf("expected mocknotify notify reuser, got %+v", got)
		}
		for _, smoke := range []struct{ provider, category string }{
			{"mockdns", "apply"},
			{"mockdns", "access"},
			{"mockdns-deploy", "deploy"},
			{"mocknotify", "notify"},
		} {
			var envelope struct {
				Data struct {
					Provider string `json:"provider"`
					Category string `json:"category"`
				} `json:"data"`
			}
			if err := call("GET", "/api/provider-schemas/"+smoke.provider+"?category="+smoke.category, nil, &envelope); err != nil {
				t.Fatalf("schema %s/%s: %v", smoke.provider, smoke.category, err)
			}
			if envelope.Data.Provider != smoke.provider || envelope.Data.Category != smoke.category {
				t.Fatalf("schema %s/%s returned %+v", smoke.provider, smoke.category, envelope)
			}
		}
	})

	create := func(collection string, body any) string {
		var result struct {
			ID string `json:"id"`
		}
		if err := call("POST", "/api/collections/"+collection+"/records", body, &result); err != nil {
			t.Fatal(err)
		}
		return result.ID
	}
	accessID := create("access", map[string]any{"name": "Mock DNS access", "provider": "mockdns", "config": map[string]any{"apiUrl": apiServer.URL, "apiKey": "plugin-dns-secret"}})
	caID := create("access", map[string]any{"name": "Plugin ACME", "provider": "acmeca", "config": map[string]any{"endpoint": ca.URL + "/directory"}})
	webhookID := create("access", map[string]any{"name": "Plugin notify webhook", "provider": "webhook", "config": map[string]any{"url": apiServer.URL + "/notify", "method": "POST", "headers": "Content-Type: application/json", "allowInsecureConnections": false}})

	runWorkflow := func(name string, nodes []any) string {
		graph := map[string]any{"nodes": nodes}
		workflowID := create("workflow", map[string]any{"name": name, "trigger": "manual", "enabled": true, "hasContent": true, "hasDraft": false, "graphContent": graph, "graphDraft": graph})
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
		deadline := time.Now().Add(90 * time.Second)
		for {
			if err := call("GET", "/api/collections/workflow_run/records/"+run.Data.RunID, nil, &result); err != nil {
				t.Fatal(err)
			}
			if result.Status == "succeeded" || result.Status == "failed" || result.Status == "canceled" {
				break
			}
			if time.Now().After(deadline) {
				result.Status = "timeout"
				break
			}
			time.Sleep(30 * time.Millisecond)
		}
		var logs struct {
			Items []struct {
				NodeId  string `json:"nodeId"`
				Message string `json:"message"`
			} `json:"items"`
		}
		_ = call("GET", "/api/collections/workflow_logs/records?perPage=200&sort=timestamp&filter="+urlQueryEscape("runRef='"+run.Data.RunID+"'"), nil, &logs)
		for _, item := range logs.Items {
			t.Logf("run log [%s]: %s", item.NodeId, item.Message)
		}
		if result.Status != "succeeded" {
			f.mu.Lock()
			value, expected, issued, txt := f.value, f.expected, f.issued, f.txtResponses
			f.mu.Unlock()
			api.mu.Lock()
			presents, cleanups := api.presents, api.cleanups
			api.mu.Unlock()
			t.Logf("fixture state: presents=%d cleanups=%d value=%q expected=%q issued=%d txtResponses=%d", presents, cleanups, value, expected, issued, txt)
			t.Fatalf("workflow %s: %s %s", name, result.Status, result.Error)
		}
		return run.Data.RunID
	}
	node := func(id, kind string, config map[string]any) map[string]any {
		return map[string]any{"id": id, "type": kind, "data": map[string]any{"name": id, "config": config}}
	}

	t.Run("issue-and-deploy", func(t *testing.T) {
		applyConfig := map[string]any{"domains": f.domain, "contactEmail": "plugin@example.test", "challengeType": "dns-01", "provider": "mockdns", "providerAccessId": accessID, "caProvider": "acmeca", "caProviderAccessId": caID, "keyAlgorithm": "EC256", "nameservers": udp.LocalAddr().String(), "dnsPropagationTimeout": 10, "dnsTTL": 60, "disableARI": true, "disableFollowCNAME": true}
		deployConfig := map[string]any{"certificateOutputNodeId": "apply", "provider": "mockdns-deploy", "providerAccessId": accessID, "providerConfig": map[string]any{}}
		nodes := []any{node("start", "start", map[string]any{}), node("apply", "bizApply", applyConfig), node("deploy", "bizDeploy", deployConfig), node("end", "end", map[string]any{})}
		runID := runWorkflow("Plugin DNS issue+deploy", nodes)

		f.mu.Lock()
		chain := f.chain
		txtResponses := f.txtResponses
		f.mu.Unlock()
		api.mu.Lock()
		presents, cleanups, deploys, deployedCert := api.presents, api.cleanups, api.deploys, api.deployedCert
		api.mu.Unlock()
		if presents != 1 || cleanups != 1 || deploys != 1 {
			t.Fatalf("plugin calls presents=%d cleanups=%d deploys=%d", presents, cleanups, deploys)
		}
		if txtResponses == 0 {
			t.Fatal("no DNS query received the expected TXT value")
		}
		if strings.TrimSpace(deployedCert) != strings.TrimSpace(string(chain)) {
			t.Fatal("deploy did not receive the issued certificate")
		}
		var certificates struct {
			Items []struct {
				Certificate string `json:"certificate"`
				PrivateKey  string `json:"privateKey"`
			} `json:"items"`
		}
		if err := call("GET", "/api/collections/certificate/records?perPage=100", nil, &certificates); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, record := range certificates.Items {
			if strings.TrimSpace(record.Certificate) == strings.TrimSpace(string(chain)) {
				leaf, _ := pem.Decode([]byte(record.Certificate))
				if leaf == nil {
					t.Fatal("persisted certificate not PEM")
				}
				parsed, err := x509.ParseCertificate(leaf.Bytes)
				if err != nil || parsed.VerifyHostname(f.domain) != nil {
					t.Fatal("persisted certificate SAN mismatch")
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
		t.Logf("workflow %s issued via mockdns plugin, deployed via mockdns-deploy reuser, TXT served in %d responses", runID, txtResponses)
	})

	t.Run("notify", func(t *testing.T) {
		notifyConfig := map[string]any{"provider": "mocknotify", "providerAccessId": webhookID, "subject": "Plugin notify subject", "message": "Plugin notify message"}
		nodes := []any{node("start", "start", map[string]any{}), node("notify", "bizNotify", notifyConfig), node("end", "end", map[string]any{})}
		runWorkflow("Plugin notify", nodes)

		api.mu.Lock()
		notifies, subject, message := api.notifies, api.subject, api.message
		api.mu.Unlock()
		if notifies != 1 {
			t.Fatalf("notify calls=%d", notifies)
		}
		if subject != "Plugin notify subject" || message != "Plugin notify message" {
			t.Fatalf("notify payload subject=%q message=%q", subject, message)
		}
	})
}
