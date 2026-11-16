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
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestExternalDNSPluginsWorkflow(t *testing.T) {
	cloudPlugins := []struct {
		name string
		icon string
		run  func(*testing.T, string, string)
	}{
		{name: "cmcccloud-dns", icon: "cmcccloud.svg", run: runCMCCCloudDNSWorkflow},
		{name: "googlecloud-dns", icon: "google.svg", run: runGoogleCloudDNSWorkflow},
	}
	if binary := os.Getenv("CERTIMATE_EXTERNAL_DNS_CORE"); binary != "" {
		name := os.Getenv("CERTIMATE_EXTERNAL_DNS_PROVIDER")
		for _, provider := range cloudPlugins {
			if provider.name == name {
				provider.run(t, binary, os.Getenv("CERTIMATE_EXTERNAL_DNS_PLUGINS"))
				return
			}
		}
		t.Fatalf("unknown DNS plugin fixture %q", name)
	}
	pluginRoot := os.Getenv("CERTIMATE_EXTERNAL_DNS_PLUGIN_ROOT")
	if pluginRoot == "" {
		pluginRoot = filepath.Join("..", "..", "..", "plugins")
	}
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
		configure func(*testing.T, *b3DNSFixture, bool) b3DNSProviderConfig
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
			var manifest struct {
				Icon string `json:"icon"`
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			icon, err := os.ReadFile(filepath.Join(source, manifest.Icon))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(stage, manifest.Icon), icon, 0o644); err != nil {
				t.Fatal(err)
			}

			runDNSWorkflow(t, core, provider.name, dnsWorkflowPlugins{
				dir:              plugins,
				accessByProvider: map[string]string{provider.name: provider.name},
				configureDNS: func(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
					return provider.configure(t, f, false)
				},
				wildcardOrder: true,
			})
			t.Run("former-builtin", func(t *testing.T) {
				runDNSWorkflow(t, core, provider.name, dnsWorkflowPlugins{
					dir:              plugins,
					accessType:       strings.TrimSuffix(provider.name, "-dns"),
					pollingInterval:  1,
					wildcardOrder:    provider.name == "rfc2136-dns",
					accessByProvider: map[string]string{provider.name: provider.name},
					configureDNS: func(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
						return provider.configure(t, f, true)
					},
				})
			})
		})
	}
	for _, provider := range cloudPlugins {
		t.Run(provider.name, func(t *testing.T) {
			stage := t.TempDir()
			pluginDir := filepath.Join(stage, "plugins")
			destination := filepath.Join(pluginDir, provider.name)
			if err := os.MkdirAll(destination, 0o755); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(pluginRoot, provider.name)
			for _, name := range []string{"manifest.json", provider.icon} {
				data, err := os.ReadFile(filepath.Join(source, name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(destination, name), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("go", "build", "-o", filepath.Join(destination, provider.name), ".")
			cmd.Dir = source
			if runtime.GOOS != "linux" {
				cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
			}
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build %s plugin: %v\n%s", provider.name, err, out)
			}
			if runtime.GOOS == "linux" {
				provider.run(t, core, pluginDir)
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
					t.Fatalf("build %s workflow fixture: %v\n%s", provider.name, err, out)
				}
			}
			cmd = exec.Command("docker", "run", "--rm", "--network=none",
				"--mount", "type=bind,src="+stage+",dst=/fixture,readonly",
				"-e", "CERTIMATE_EXTERNAL_DNS_CORE=/fixture/core",
				"-e", "CERTIMATE_EXTERNAL_DNS_PLUGINS=/fixture/plugins",
				"-e", "CERTIMATE_EXTERNAL_DNS_PROVIDER="+provider.name,
				"alpine:latest", "/fixture/dns.test", "-test.run=^TestExternalDNSPluginsWorkflow$", "-test.v", "-test.timeout=4m")
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
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

func runGoogleCloudDNSWorkflow(t *testing.T, binary, pluginDir string) {
	for _, provider := range []string{"googlecloud", "googlecloud-dns"} {
		t.Run(provider, func(t *testing.T) {
			runDNSWorkflow(t, binary, "googlecloud", dnsWorkflowPlugins{
				dir:              pluginDir,
				accessByProvider: map[string]string{"googlecloud-dns": "googlecloud"},
				cname:            provider == "googlecloud",
				configureDNS: func(t *testing.T, f *b3DNSFixture) b3DNSProviderConfig {
					config := b3DNSConfig(t, f, "googlecloud")
					config.Provider = provider
					f.googleTTL = 123
					config.Apply = map[string]any{"dnsTTL": f.googleTTL, "disableFollowCNAME": f.cname == ""}
					config.Env = []string{"GCE_ZONE_ID=", "GCE_POLLING_INTERVAL=1"}
					return config
				},
			})
		})
	}
}

func configureRFC2136Fixture(t *testing.T, f *b3DNSFixture, legacy bool) b3DNSProviderConfig {
	zone := dns.Fqdn(f.domain)
	algorithm := dns.HmacSHA256
	if legacy {
		algorithm = dns.HmacSHA1
		if f.wildcard == nil {
			zone = "example.test."
		}
	}
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
			if legacy && req.Opcode == dns.OpcodeQuery {
				reply.Authoritative = true
				if len(req.Question) == 1 && req.Question[0].Qtype == dns.TypeSOA && req.Question[0].Name == zone {
					reply.Answer = []dns.RR{&dns.SOA{Hdr: dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60}, Ns: "ns.example.test.", Mbox: "admin.example.test."}}
				}
				if err := w.WriteMsg(reply); err != nil {
					t.Errorf("write DNS SOA response: %v", err)
				}
				return
			}
			signature := req.IsTsig()
			if w.TsigStatus() != nil || signature == nil || signature.Hdr.Name != key || signature.Algorithm != algorithm {
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
	if legacy {
		packet, err := net.ListenPacket("udp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		resolver := &dns.Server{PacketConn: packet, Handler: server.Handler, MsgAcceptFunc: server.MsgAcceptFunc, TsigSecret: server.TsigSecret}
		go resolver.ActivateAndServe()
		t.Cleanup(func() { _ = resolver.Shutdown() })
	}
	if legacy {
		addr := listener.Addr().(*net.TCPAddr)
		return b3DNSProviderConfig{
			Provider: "rfc2136",
			Access:   map[string]any{"host": addr.IP.String(), "port": addr.Port, "tsigKey": key, "tsigSecret": secret},
			Apply:    map[string]any{"dnsTTL": 45},
		}
	}
	return b3DNSProviderConfig{
		Provider: "rfc2136-dns",
		Access:   map[string]any{"server": listener.Addr().String(), "zone": f.domain, "keyName": key, "keySecret": secret},
		Apply:    map[string]any{"dnsTTL": 1, "providerConfig": map[string]any{"ttl": 45}},
	}
}

func configurePowerDNSFixture(t *testing.T, f *b3DNSFixture, legacy bool) b3DNSProviderConfig {
	zone := dns.Fqdn(f.domain)
	if legacy {
		zone = "example.test."
	}
	path := "/api/v1/servers/localhost/zones/" + zone
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if legacy && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api":
				_, _ = w.Write([]byte(`[{"version":1}]`))
			case path:
				rrsets := []any{}
				if f.value != "" {
					rrsets = append(rrsets, map[string]any{"name": f.challengeRecordName(), "type": "TXT", "records": []any{map[string]any{"content": strconv.Quote(f.value)}}})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"id": zone, "kind": "Native", "rrsets": rrsets})
			default:
				t.Errorf("unexpected PowerDNS read: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}
		if r.Method != http.MethodPatch || r.URL.Path != path || r.Header.Get("X-API-Key") != "fixture-api-key" {
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
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.RRSets) != 1 || body.RRSets[0].Type != "TXT" {
			t.Errorf("invalid PowerDNS payload: %+v, %v", body, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		set := body.RRSets[0]
		if legacy && set.ChangeType == "DELETE" && len(set.Records) == 0 {
			applyExternalDNSUpdate(t, f, set.Name, f.value, false)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if len(set.Records) != 1 {
			t.Errorf("invalid PowerDNS record set: %+v", set)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		value, err := strconv.Unquote(set.Records[0].Content)
		if err != nil || (set.ChangeType != "EXTEND" && set.ChangeType != "PRUNE" && !(legacy && set.ChangeType == "REPLACE")) {
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
		applyExternalDNSUpdate(t, f, set.Name, value, set.ChangeType != "PRUNE")
		w.WriteHeader(http.StatusNoContent)
	}))
	if legacy {
		server.StartTLS()
	} else {
		server.Start()
	}
	t.Cleanup(server.Close)
	if legacy {
		return b3DNSProviderConfig{
			Provider: "powerdns",
			Access:   map[string]any{"serverUrl": server.URL, "apiKey": "fixture-api-key", "allowInsecureConnections": true},
			Apply:    map[string]any{"dnsTTL": 45},
		}
	}
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
