package certacme_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestExternalDNSPluginsWorkflow(t *testing.T) {
	if binary := os.Getenv("CERTIMATE_EXTERNAL_DNS_CMCC_CORE"); binary != "" {
		runCMCCCloudDNSWorkflow(t, binary, os.Getenv("CERTIMATE_EXTERNAL_DNS_CMCC_PLUGINS"))
		return
	}
	pluginRoot := filepath.Join("..", "..", "..", "plugins")
	if _, err := os.Stat(filepath.Join(pluginRoot, "go.mod")); os.IsNotExist(err) {
		t.Skip("requires the sibling plugins checkout")
	} else if err != nil {
		t.Fatal(err)
	}
	core := filepath.Join(t.TempDir(), "certimate")
	cmd := exec.Command("go", "build", "-o", core, ".")
	cmd.Dir = filepath.Join("..", "..")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build Certimate: %v\n%s", err, output)
	}

	for _, provider := range []struct {
		name      string
		configure func(*testing.T, *b3DNSFixture) b3DNSProviderConfig
	}{
		{name: "rfc2136-dns", configure: configureRFC2136Fixture},
		{name: "powerdns-dns", configure: configurePowerDNSFixture},
	} {
		t.Run(provider.name, func(t *testing.T) {
			plugins := t.TempDir()
			stage := filepath.Join(plugins, provider.name)
			if err := os.Mkdir(stage, 0o755); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(pluginRoot, provider.name)
			binary, err := filepath.Abs(filepath.Join(stage, provider.name))
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "build", "-o", binary, ".")
			cmd.Dir = source
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build %s: %v\n%s", provider.name, err, output)
			}
			data, err := os.ReadFile(filepath.Join(source, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(stage, "manifest.json"), data, 0o644); err != nil {
				t.Fatal(err)
			}

			runDNSWorkflow(t, core, provider.name, dnsWorkflowPlugins{
				dir:              plugins,
				accessByProvider: map[string]string{provider.name: provider.name},
				configureDNS:     provider.configure,
				wildcardOrder:    true,
			})
		})
	}
	t.Run("cmcccloud-dns", func(t *testing.T) {
		stage := t.TempDir()
		pluginDir := filepath.Join(stage, "plugins")
		destination := filepath.Join(pluginDir, "cmcccloud-dns")
		if err := os.MkdirAll(destination, 0o755); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(pluginRoot, "cmcccloud-dns")
		for _, name := range []string{"manifest.json", "cmcccloud.svg"} {
			data, err := os.ReadFile(filepath.Join(source, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(destination, name), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		plugin := filepath.Join(destination, "cmcccloud-dns")
		cmd := exec.Command("go", "build", "-o", plugin, ".")
		cmd.Dir = source
		if runtime.GOOS != "linux" {
			cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build CMCC Cloud DNS plugin: %v\n%s", err, out)
		}
		if runtime.GOOS == "linux" {
			runCMCCCloudDNSWorkflow(t, core, pluginDir)
			return
		}
		for _, build := range []struct {
			dir  string
			args []string
		}{
			{dir: "../..", args: []string{"build", "-o", filepath.Join(stage, "core"), "."}},
			{dir: ".", args: []string{"test", "-c", "-o", filepath.Join(stage, "dns.test"), "."}},
		} {
			cmd := exec.Command("go", build.args...)
			cmd.Dir = build.dir
			cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build CMCC Cloud DNS workflow fixture: %v\n%s", err, out)
			}
		}
		cmd = exec.Command("docker", "run", "--rm", "--network=none",
			"--mount", "type=bind,src="+stage+",dst=/fixture,readonly",
			"-e", "CERTIMATE_EXTERNAL_DNS_CMCC_CORE=/fixture/core",
			"-e", "CERTIMATE_EXTERNAL_DNS_CMCC_PLUGINS=/fixture/plugins",
			"alpine:latest", "/fixture/dns.test", "-test.run=^TestExternalDNSPluginsWorkflow$", "-test.v", "-test.timeout=4m")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func runCMCCCloudDNSWorkflow(t *testing.T, binary, pluginDir string) {
	for _, provider := range []string{"cmcccloud", "cmcccloud-dns"} {
		t.Run(provider, func(t *testing.T) {
			runDNSWorkflow(t, binary, "cmcccloud", dnsWorkflowPlugins{
				dir:              pluginDir,
				accessByProvider: map[string]string{"cmcccloud-dns": "cmcccloud"},
				wildcardOrder:    true,
				cname:            provider == "cmcccloud",
				configureDNS: func(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
					config, _ := b6B8DNSConfig(t, f)
					config.Provider = provider
					config.Apply = map[string]any{"dnsTTL": 123, "disableFollowCNAME": f.cname == ""}
					return config
				},
			})
		})
	}
}

func configureRFC2136Fixture(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
	zone := f.domain
	const key = "certimate.example.test."
	const secret = "Zml4dHVyZS10c2lnLXNlY3JldA=="
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{
		Listener:   listener,
		TsigSecret: map[string]string{key: secret},
		MsgAcceptFunc: func(dns.Header) dns.MsgAcceptAction {
			return dns.MsgAccept
		},
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
			f.mu.Lock()
			defer f.mu.Unlock()
			reply := new(dns.Msg).SetReply(req)
			signature := req.IsTsig()
			if w.TsigStatus() != nil || signature == nil || signature.Hdr.Name != key || signature.Algorithm != dns.HmacSHA256 {
				t.Error("DNS update has invalid TSIG authentication")
				reply.Rcode = dns.RcodeNotAuth
			} else {
				reply.SetTsig(key, signature.Algorithm, 300, time.Now().Unix())
				if f.mode == "auth-failure" {
					f.denied++
					reply.Rcode = dns.RcodeRefused
				} else if req.Opcode != dns.OpcodeUpdate || len(req.Question) != 1 || req.Question[0].Name != dns.Fqdn(zone) || len(req.Ns) != 1 {
					t.Errorf("invalid DNS update: %s", req)
					reply.Rcode = dns.RcodeFormatError
				} else if txt, ok := req.Ns[0].(*dns.TXT); !ok || len(txt.Txt) != 1 {
					t.Errorf("invalid DNS TXT update: %s", req)
					reply.Rcode = dns.RcodeFormatError
				} else {
					switch txt.Hdr.Class {
					case dns.ClassINET:
						if txt.Hdr.Ttl != 45 {
							t.Errorf("TXT TTL = %d, want 45", txt.Hdr.Ttl)
						}
						if rejectChallengeWrite(f, txt.Txt[0]) {
							reply.Rcode = dns.RcodeRefused
							break
						}
						applyExternalDNSUpdate(t, f, txt.Hdr.Name, txt.Txt[0], true)
					case dns.ClassNONE:
						applyExternalDNSUpdate(t, f, txt.Hdr.Name, txt.Txt[0], false)
					default:
						t.Errorf("unexpected TXT update class %d", txt.Hdr.Class)
						reply.Rcode = dns.RcodeFormatError
					}
				}
			}
			if err := w.WriteMsg(reply); err != nil {
				t.Errorf("write DNS update response: %v", err)
			}
		}),
	}
	go server.ActivateAndServe()
	t.Cleanup(func() { _ = server.Shutdown() })
	return b3DNSProviderConfig{
		Provider: "rfc2136-dns",
		Access:   map[string]any{"server": listener.Addr().String(), "zone": f.domain, "keyName": key, "keySecret": secret},
		Apply:    map[string]any{"dnsTTL": 1, "providerConfig": map[string]any{"ttl": 45}},
	}
}

func configurePowerDNSFixture(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/servers/localhost/zones/"+dns.Fqdn(f.domain) || r.Header.Get("X-API-Key") != "fixture-api-key" {
			t.Error("invalid PowerDNS request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if f.mode == "auth-failure" {
			f.denied++
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body struct {
			RRSets []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				ChangeType string `json:"changetype"`
				TTL        uint32 `json:"ttl"`
				Records    []struct {
					Content string `json:"content"`
				} `json:"records"`
			} `json:"rrsets"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.RRSets) != 1 || body.RRSets[0].Type != "TXT" || len(body.RRSets[0].Records) != 1 {
			t.Errorf("invalid PowerDNS payload: %+v, %v", body, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		set := body.RRSets[0]
		value, err := strconv.Unquote(set.Records[0].Content)
		if err != nil || (set.ChangeType != "EXTEND" && set.ChangeType != "PRUNE") {
			t.Errorf("invalid PowerDNS update: %+v, %v", set, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if set.TTL != 45 {
			t.Errorf("TXT TTL = %d, want 45", set.TTL)
		}
		if set.ChangeType == "EXTEND" && rejectChallengeWrite(f, value) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		applyExternalDNSUpdate(t, f, set.Name, value, set.ChangeType == "EXTEND")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return b3DNSProviderConfig{
		Provider: "powerdns-dns",
		Access:   map[string]any{"serverUrl": server.URL, "zone": f.domain, "apiKey": "fixture-api-key"},
		Apply:    map[string]any{"dnsTTL": 1, "providerConfig": map[string]any{"ttl": 45}},
	}
}

func rejectChallengeWrite(f *b3DNSFixture, value string) bool {
	if f.mode != "partial-present-failure" || value != f.wildcard.expected {
		return false
	}
	f.denied++
	return true
}

func applyExternalDNSUpdate(t *testing.T, f *b3DNSFixture, name, value string, present bool) {
	record := &f.value
	if f.wildcard != nil && value == f.wildcard.expected {
		record = &f.wildcard.value
	} else if value != f.expected {
		t.Errorf("unexpected challenge TXT value %q", value)
		return
	}
	if name != f.challengeRecordName() {
		t.Errorf("unexpected challenge TXT name=%q value=%q", name, value)
	}
	if present {
		f.creates++
		*record = value
	} else if *record == value {
		f.deletes++
		*record = ""
	}
}
