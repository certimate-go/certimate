package certacme_test

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestB5DNSWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_B5_DNS_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_B5_DNS_CORE to the Linux core executable")
	}
	pluginDir := os.Getenv("CERTIMATE_B5_DNS_PLUGINS")
	if os.Getenv("CERTIMATE_B5_DNS_CONTAINER") == "" {
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
		args := []string{"run", "--rm", "--network", "none", "-v", dir + ":/fixture:ro", "-v", binary + ":/core:ro", "-e", "CERTIMATE_B5_DNS_CORE=/core", "-e", "CERTIMATE_B5_DNS_CONTAINER=1"}
		if pluginDir != "" {
			pluginDir, err = filepath.Abs(pluginDir)
			if err != nil {
				t.Fatal(err)
			}
			args = append(args, "-v", pluginDir+":/plugins:ro", "-e", "CERTIMATE_B5_DNS_PLUGINS=/plugins")
		}
		cmd = exec.Command("docker", append(args, "alpine:latest", "/fixture/dns.test", "-test.run=^TestB5DNSWorkflow$", "-test.v", "-test.timeout=5m")...)
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
			"jdcloud-alb": "jdcloud", "jdcloud-cdn": "jdcloud", "jdcloud-live": "jdcloud", "jdcloud-ssl": "jdcloud", "jdcloud-vod": "jdcloud", "jdcloud-waf": "jdcloud",
			"ucloud-ualb": "ucloud", "ucloud-ucdn": "ucloud", "ucloud-uclb": "ucloud", "ucloud-uewaf": "ucloud", "ucloud-pathx": "ucloud", "ucloud-us3": "ucloud",
		}
	}
	for _, family := range []string{"jdcloud", "ucloud"} {
		t.Run(family, func(t *testing.T) { runDNSWorkflow(t, binary, family, plugins) })
	}
}

func b5DNSConfig(f *b3DNSFixture) (b3DNSProviderConfig, bool) {
	switch f.family {
	case "jdcloud":
		f.proxyHosts = append(f.proxyHosts, "domainservice.jdcloud-api.com")
		return b3DNSProviderConfig{Provider: "jdcloud-dns", Access: map[string]any{"accessKeyId": "fixture-dns-id", "accessKeySecret": "fixture-dns-secret"}}, true
	case "ucloud":
		f.proxyHosts = append(f.proxyHosts, "api.ucloud.cn")
		return b3DNSProviderConfig{Provider: "ucloud-udnr", Access: map[string]any{"publicKey": "fixture-dns-id", "privateKey": "fixture-dns-secret", "projectId": "fixture-project"}}, true
	}
	return b3DNSProviderConfig{}, false
}

func (f *b3DNSFixture) b5DNSCloud(w http.ResponseWriter, r *http.Request) bool {
	if f.family != "jdcloud" && f.family != "ucloud" {
		return false
	}
	reply := func(v any) { json.NewEncoder(w).Encode(v) }
	fail := func(reason string) {
		f.t.Errorf("%s DNS fixture %s %s: %s", f.family, r.Method, r.URL.Path, reason)
		w.WriteHeader(http.StatusBadRequest)
		reply(map[string]string{"message": reason})
	}
	if f.family == "jdcloud" {
		if r.Host != "domainservice.jdcloud-api.com" || !strings.Contains(r.Header.Get("Authorization"), "Credential=fixture-dns-id/") {
			fail("unexpected host or JDCloud credential")
			return true
		}
		const domainPath = "/v2/regions/cn-north-1/domain"
		switch {
		case r.Method == "GET" && r.URL.Path == domainPath:
			reply(map[string]any{"result": map[string]any{"dataList": []any{map[string]any{"id": 42, "domainName": "example.test"}}, "totalPage": 1}})
		case r.Method == "POST" && r.URL.Path == domainPath+"/42/ResourceRecord":
			var body struct {
				Req struct {
					HostRecord string `json:"hostRecord"`
					HostValue  string `json:"hostValue"`
					Type       string `json:"type"`
				} `json:"req"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				fail("invalid create body")
				break
			}
			name := strings.TrimSuffix("_acme-challenge."+f.domain, ".example.test")
			if f.value != "" || body.Req.HostRecord != name || body.Req.Type != "TXT" || body.Req.HostValue != f.expected {
				fail("wrong TXT creation")
				break
			}
			f.value = body.Req.HostValue
			f.creates++
			reply(map[string]any{"result": map[string]any{"dataList": map[string]int{"id": 7}}})
		case r.Method == "DELETE" && r.URL.Path == domainPath+"/42/ResourceRecord/7":
			if f.value == "" {
				fail("TXT record already absent")
				break
			}
			f.value = ""
			f.deletes++
			reply(map[string]any{"result": map[string]any{}})
		default:
			fail("unexpected JDCloud DNS route")
		}
		return true
	}
	if err := r.ParseForm(); err != nil {
		fail("invalid UCloud request form")
		return true
	}
	keys := make([]string, 0, len(r.Form))
	for key := range r.Form {
		if key != "Signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var signed strings.Builder
	for _, key := range keys {
		signed.WriteString(key)
		signed.WriteString(r.Form.Get(key))
	}
	signed.WriteString("fixture-dns-secret")
	digest := sha1.Sum([]byte(signed.String()))
	if r.Host != "api.ucloud.cn" || r.Form.Get("PublicKey") != "fixture-dns-id" || r.Form.Get("ProjectId") != "fixture-project" || r.Form.Get("Signature") != hex.EncodeToString(digest[:]) {
		fail("unexpected host, project or UCloud signature")
		return true
	}
	if r.Form.Get("Dn") != "example.test" {
		fail("unexpected DNS zone")
		return true
	}
	name := "_acme-challenge." + f.domain
	switch r.Form.Get("Action") {
	case "UdnrDomainDNSAdd":
		if f.value != "" || r.Form.Get("RecordName") != name || r.Form.Get("DnsType") != "TXT" || r.Form.Get("Content") != f.expected {
			fail("wrong TXT creation")
			break
		}
		f.value = r.Form.Get("Content")
		f.creates++
		reply(map[string]any{"RetCode": 0})
	case "UdnrDomainDNSQuery":
		data := []map[string]string{}
		if f.value != "" {
			data = append(data, map[string]string{"RecordName": name, "DnsType": "TXT", "Content": f.value})
		}
		reply(map[string]any{"RetCode": 0, "Data": data})
	case "UdnrDeleteDnsRecord":
		if f.value == "" || r.Form.Get("RecordName") != name || r.Form.Get("DnsType") != "TXT" || r.Form.Get("Content") != f.value {
			fail("wrong TXT deletion")
			break
		}
		f.value = ""
		f.deletes++
		reply(map[string]any{"RetCode": 0})
	default:
		fail("unexpected UCloud action")
	}
	return true
}
