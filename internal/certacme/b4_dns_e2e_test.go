package certacme_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestB4DNSWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_B4_DNS_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_B4_DNS_CORE to the Linux core executable")
	}
	pluginDir := os.Getenv("CERTIMATE_B4_DNS_PLUGINS")
	if os.Getenv("CERTIMATE_B4_DNS_CONTAINER") == "" {
		dir := t.TempDir()
		cmd := exec.Command("go", "test", "-c", "-o", filepath.Join(dir, "dns.test"), ".")
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build DNS test: %v %s", err, out)
		}
		binary, err := filepath.Abs(binary)
		if err != nil {
			t.Fatal(err)
		}
		args := []string{
			"run", "--rm", "--network", "none",
			"--add-host", "dns.baidubce.com:127.0.0.1",
			"-v", dir + ":/fixture:ro", "-v", binary + ":/core:ro",
			"-e", "CERTIMATE_B4_DNS_CORE=/core", "-e", "CERTIMATE_B4_DNS_CONTAINER=1",
		}
		if pluginDir != "" {
			pluginDir, err = filepath.Abs(pluginDir)
			if err != nil {
				t.Fatal(err)
			}
			args = append(args, "-v", pluginDir+":/plugins:ro", "-e", "CERTIMATE_B4_DNS_PLUGINS=/plugins")
		}
		cmd = exec.Command("docker", append(args, "alpine:latest", "/fixture/dns.test", "-test.run=^TestB4DNSWorkflow$", "-test.v", "-test.timeout=5m")...)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	plugins := dnsWorkflowPlugins{dir: pluginDir}
	if pluginDir != "" {
		plugins.accessByProvider = map[string]string{
			"rainyun-rcdn": "rainyun", "rainyun-sslcenter": "rainyun",
			"baiducloud-appblb": "baiducloud", "baiducloud-blb": "baiducloud",
			"baiducloud-cdn": "baiducloud", "baiducloud-cert": "baiducloud",
		}
	}
	for _, family := range []string{"rainyun", "baiducloud"} {
		t.Run(family, func(t *testing.T) { runDNSWorkflow(t, binary, family, plugins) })
	}
}

func b4DNSConfig(t *testing.T, f *b3DNSFixture) (b3DNSProviderConfig, bool) {
	t.Helper()
	switch f.family {
	case "rainyun":
		f.proxyHosts = append(f.proxyHosts, "api.v2.rainyun.com")
		return b3DNSProviderConfig{Provider: "rainyun", Access: map[string]any{"apiKey": "fixture-dns-token"}}, true
	case "baiducloud":
		listener, err := net.Listen("tcp", "127.0.0.1:80")
		if err != nil {
			t.Fatal(err)
		}
		target := &http.Server{Handler: http.HandlerFunc(f.cloud)}
		go target.Serve(listener)
		t.Cleanup(func() { target.Close() })
		return b3DNSProviderConfig{Provider: "baiducloud-dns", Access: map[string]any{"accessKeyId": "fixture-dns-id", "secretAccessKey": "fixture-dns-secret"}}, true
	}
	return b3DNSProviderConfig{}, false
}

func (f *b3DNSFixture) b4DNSCloud(w http.ResponseWriter, r *http.Request) bool {
	const (
		rainRecordPath  = "/product/domain/42/dns"
		baiduRecordPath = "/v1/dns/zone/example.test/record"
	)
	name := strings.TrimSuffix("_acme-challenge."+f.domain, ".example.test")
	records := func(record map[string]any) []map[string]any {
		if f.value == "" {
			return []map[string]any{}
		}
		return []map[string]any{record}
	}
	rainList := func(records []map[string]any) map[string]any {
		return map[string]any{"code": 200, "data": map[string]any{"Records": records, "TotalRecords": len(records)}}
	}
	reply := func(v any) { json.NewEncoder(w).Encode(v) }
	fail := func(status int, reason string) {
		f.t.Errorf("%s DNS fixture %s %s: %s", f.family, r.Method, r.URL.Path, reason)
		w.WriteHeader(status)
		reply(map[string]string{"message": reason})
	}
	create := func(recordName, kind, value string) {
		if f.value != "" || recordName != name || kind != "TXT" || value != f.expected {
			fail(http.StatusBadRequest, "invalid TXT record")
			return
		}
		f.value = value
		f.creates++
		reply(map[string]int{"code": 200})
	}
	remove := func() {
		if f.value == "" {
			fail(http.StatusNotFound, "TXT record already absent")
			return
		}
		f.value = ""
		f.deletes++
		reply(map[string]int{"code": 200})
	}
	switch f.family {
	case "rainyun":
		if r.Host != "api.v2.rainyun.com" || r.Header.Get("X-API-Key") != "fixture-dns-token" {
			fail(http.StatusForbidden, "unexpected host or API key")
			return true
		}
		switch {
		case r.Method == "GET" && r.URL.Path == "/product/domain":
			reply(rainList([]map[string]any{{"id": 42, "domain": "example.test"}}))
		case r.Method == "GET" && r.URL.Path == rainRecordPath:
			reply(rainList(records(map[string]any{"record_id": 7, "host": name, "type": "TXT", "value": f.value})))
		case r.Method == "POST" && r.URL.Path == rainRecordPath:
			var body struct{ Host, Type, Value string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				fail(http.StatusBadRequest, fmt.Sprintf("decode RainYun TXT record: %v", err))
				break
			}
			create(body.Host, body.Type, body.Value)
		case r.Method == "DELETE" && r.URL.Path == rainRecordPath && r.URL.Query().Get("record_id") == "7":
			remove()
		default:
			fail(http.StatusNotFound, "unexpected RainYun DNS route")
		}
	case "baiducloud":
		if r.Host != "dns.baidubce.com" || !strings.HasPrefix(r.Header.Get("Authorization"), "bce-auth-v1/fixture-dns-id/") {
			fail(http.StatusForbidden, "unexpected host or BCE credential")
			return true
		}
		switch {
		case r.Method == "GET" && r.URL.Path == baiduRecordPath:
			reply(map[string]any{"records": records(map[string]any{"id": "record-7", "rr": name, "type": "TXT", "value": f.value})})
		case r.Method == "POST" && r.URL.Path == baiduRecordPath:
			var body struct{ RR, Type, Value string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				fail(http.StatusBadRequest, fmt.Sprintf("decode Baidu TXT record: %v", err))
				break
			}
			create(body.RR, body.Type, body.Value)
		case r.Method == "DELETE" && r.URL.Path == baiduRecordPath+"/record-7":
			remove()
		default:
			fail(http.StatusNotFound, "unexpected Baidu DNS route")
		}
	default:
		return false
	}
	return true
}
