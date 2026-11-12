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
	"strconv"
	"strings"
	"testing"
)

const b6B8HuaweiIAM = "iam.cn-north-1.myhuaweicloud.com"

var b6B8DNSHosts = map[string]string{
	"ctcccloud":   "smartdns-global.ctapi.ctyun.cn",
	"cmcccloud":   "ecloud.10086.cn",
	"huaweicloud": "dns.cn-north-1.myhuaweicloud.com",
	"byteplus":    "open.byteplusapi.com",
	"volcengine":  "open.volcengineapi.com",
}

func TestB6B8DNSWorkflow(t *testing.T) {
	binary := os.Getenv("CERTIMATE_B6_B8_DNS_CORE")
	if binary == "" {
		t.Skip("set CERTIMATE_B6_B8_DNS_CORE to the Linux core executable")
	}
	pluginDir := os.Getenv("CERTIMATE_B6_B8_DNS_PLUGINS")
	if os.Getenv("CERTIMATE_B6_B8_DNS_CONTAINER") != "1" {
		dir := t.TempDir()
		cmd := exec.Command("go", "test", "-c", "-o", filepath.Join(dir, "dns.test"), ".")
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build DNS test: %v\n%s", err, out)
		}
		binary, err := filepath.Abs(binary)
		if err != nil {
			t.Fatal(err)
		}
		args := []string{
			"run", "--rm", "--network=none",
			"--mount", "type=bind,src=" + dir + ",dst=/fixture,readonly",
			"--mount", "type=bind,src=" + binary + ",dst=/core,readonly",
			"-e", "CERTIMATE_B6_B8_DNS_CORE=/core",
			"-e", "CERTIMATE_B6_B8_DNS_CONTAINER=1",
		}
		for _, host := range []string{b6B8DNSHosts["huaweicloud"], b6B8HuaweiIAM, b6B8DNSHosts["byteplus"], b6B8DNSHosts["volcengine"]} {
			args = append(args, "--add-host", host+":127.0.0.1")
		}
		if pluginDir != "" {
			pluginDir, err = filepath.Abs(pluginDir)
			if err != nil {
				t.Fatal(err)
			}
			args = append(args, "--mount", "type=bind,src="+pluginDir+",dst=/plugins,readonly", "-e", "CERTIMATE_B6_B8_DNS_PLUGINS=/plugins")
		}
		cmd = exec.Command("docker", append(args, "alpine:latest", "/fixture/dns.test", "-test.run=^TestB6B8DNSWorkflow$", "-test.v", "-test.timeout=8m")...)
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, family := range []string{"ctcccloud", "cmcccloud", "huaweicloud", "byteplus", "volcengine"} {
		t.Run(family, func(t *testing.T) {
			var plugins dnsWorkflowPlugins
			if pluginDir != "" {
				selected := map[string]string{
					"ctcccloud": "ctcccloud-cms", "cmcccloud": "cmcccloud-cdn", "huaweicloud": "huaweicloud-scm", "byteplus": "byteplus-apig", "volcengine": "volcengine-apig",
				}
				provider := selected[family]
				plugins.dir = t.TempDir()
				if err := os.CopyFS(filepath.Join(plugins.dir, provider), os.DirFS(filepath.Join(pluginDir, provider))); err != nil {
					t.Fatal(err)
				}
				plugins.accessByProvider = map[string]string{provider: family}
			}
			runDNSWorkflow(t, binary, family, plugins)
		})
	}
}

func b6B8DNSConfig(t *testing.T, f *b3DNSFixture) (b3DNSProviderConfig, bool) {
	host, ok := b6B8DNSHosts[f.family]
	if !ok {
		return b3DNSProviderConfig{}, false
	}
	f.proxyHosts = append(f.proxyHosts, host)
	switch f.family {
	case "huaweicloud":
		f.proxyHosts = append(f.proxyHosts, b6B8HuaweiIAM)
		f.proxyTLSAddress = "127.0.0.1:443"
	case "volcengine":
		f.proxyTLSAddress = "127.0.0.1:443"
	case "byteplus":
		listener, err := net.Listen("tcp", "127.0.0.1:80")
		if err != nil {
			t.Fatal(err)
		}
		server := &http.Server{Handler: http.HandlerFunc(f.cloud)}
		go server.Serve(listener)
		t.Cleanup(func() { _ = server.Close() })
	}
	config := b3DNSProviderConfig{Provider: f.family + "-dns", Access: map[string]any{"accessKeyId": "fixture-dns-id"}}
	if f.family == "ctcccloud" {
		config.Provider = "ctcccloud-smartdns"
	}
	secret := "secretAccessKey"
	if f.family == "cmcccloud" {
		secret = "accessKeySecret"
	}
	config.Access[secret] = "fixture-dns-secret"
	return config, true
}

func (f *b3DNSFixture) b6B8DNSCloud(w http.ResponseWriter, r *http.Request) bool {
	host, ok := b6B8DNSHosts[f.family]
	if !ok {
		return false
	}
	reply := func(value any) { _ = json.NewEncoder(w).Encode(value) }
	reject := func(reason string) {
		f.t.Errorf("%s DNS fixture %s %s: %s", f.family, r.Method, r.URL.RequestURI(), reason)
		w.WriteHeader(http.StatusBadRequest)
		reply(map[string]string{"message": reason})
	}
	create := func(value string, result any) {
		if f.value != "" || value != f.expected {
			reject("incorrect TXT value")
			return
		}
		f.value = value
		f.creates++
		reply(result)
	}
	remove := func(result any) {
		if f.value == "" {
			reject("record is already absent")
			return
		}
		f.value = ""
		f.deletes++
		reply(result)
	}
	decode := func(value any) bool {
		if err := json.NewDecoder(r.Body).Decode(value); err != nil {
			reject(fmt.Sprintf("decode request: %v", err))
			return false
		}
		return true
	}
	if f.family == "huaweicloud" && r.Host == b6B8HuaweiIAM {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/projects" || r.URL.Query().Get("name") != "cn-north-1" {
			reject("unexpected IAM project lookup")
		} else {
			reply(map[string]any{"projects": []map[string]any{{"id": "fixture-project", "name": "cn-north-1", "domain_id": "fixture-domain", "enabled": true}}})
		}
		return true
	}
	if r.Host != host {
		reject("incorrect host")
		return true
	}
	const zone = "example.test"
	fqdn := "_acme-challenge." + f.domain + "."
	name := strings.TrimSuffix(fqdn, "."+zone+".")
	switch f.family {
	case "ctcccloud":
		if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("eop-authorization"), "fixture-dns-id Headers=") {
			reject("expected signed POST")
			break
		}
		switch r.URL.Path {
		case "/v2/addRecord":
			var body struct{ Domain, Host, Type, Value string }
			if !decode(&body) {
				break
			}
			if body.Domain != zone || body.Host != name || body.Type != "TXT" {
				reject("incorrect record target")
				break
			}
			create(body.Value, map[string]any{"statusCode": 200, "returnObj": map[string]int{"recordId": 7}})
		case "/v2/deleteRecord":
			var body struct{ RecordID int }
			if !decode(&body) {
				break
			}
			if body.RecordID != 7 {
				reject("incorrect record ID")
				break
			}
			remove(map[string]int{"statusCode": 200})
		default:
			reject("unexpected CTCC route")
		}
	case "cmcccloud":
		var body struct {
			Api, AccessKey, Product string
			BodyParameter           struct {
				Rr, DomainName, Type, Value string
				RecordIDList                []string
			}
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/query/openapi/apim/request/sdk" {
			reject("unexpected CMCC route")
			break
		}
		if !decode(&body) {
			break
		}
		if body.AccessKey != "fixture-dns-id" || body.Product != "clouddns" {
			reject("incorrect CMCC account or product")
			break
		}
		cmccReply := func(value any) map[string]any {
			encoded, err := json.Marshal(value)
			if err != nil {
				f.t.Fatal(err)
			}
			return map[string]any{"body": map[string]string{"responseBody": string(encoded)}}
		}
		record := body.BodyParameter
		switch body.Api {
		case "createRecordOpenapi":
			if record.Rr != name || record.DomainName != zone || record.Type != "TXT" {
				reject("incorrect record target")
				break
			}
			create(record.Value, cmccReply(map[string]any{"state": "OK", "body": map[string]string{"recordId": "record-7"}}))
		case "deleteRecordOpenapi":
			if len(record.RecordIDList) != 1 || record.RecordIDList[0] != "record-7" {
				reject("incorrect record ID")
				break
			}
			remove(cmccReply(map[string]string{"state": "OK"}))
		default:
			reject("unexpected CMCC operation")
		}
	case "huaweicloud":
		const records = "/v2/zones/zone-42/recordsets"
		if !strings.Contains(r.Header.Get("Authorization"), "Access=fixture-dns-id") {
			reject("expected signed Huawei request")
			break
		}
		switch r.Method + " " + r.URL.Path {
		case "GET /v2/zones":
			reply(map[string]any{"zones": []map[string]string{{"id": "zone-42", "name": zone + "."}}})
		case "GET " + records:
			reply(map[string]any{"recordsets": []any{}})
		case "POST " + records:
			var body struct {
				Name, Type string
				Records    []string
			}
			if !decode(&body) {
				break
			}
			if body.Name != fqdn || body.Type != "TXT" || len(body.Records) != 1 || body.Records[0] != strconv.Quote(f.expected) {
				reject("incorrect record target or value")
				break
			}
			create(f.expected, map[string]string{"id": "record-7"})
		case "GET " + records + "/record-7":
			reply(map[string]string{"id": "record-7", "status": "ACTIVE"})
		case "DELETE " + records + "/record-7":
			remove(map[string]string{"id": "record-7", "status": "DELETED"})
		default:
			reject("unexpected Huawei route")
		}
	case "byteplus", "volcengine":
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=fixture-dns-id/") || r.URL.Path != "/" {
			reject("expected signed DNS request")
			break
		}
		result := func(data any) map[string]any {
			return map[string]any{"ResponseMetadata": map[string]string{"RequestId": "request-1"}, "Result": data}
		}
		switch r.Method + " " + r.URL.Query().Get("Action") {
		case "GET ListZones":
			if r.URL.Query().Get("Key") != zone {
				reply(result(map[string]int{"Total": 0}))
			} else {
				reply(result(map[string]any{"Total": 1, "Zones": []map[string]any{{"ZID": 42, "ZoneName": zone}}}))
			}
		case "POST CreateRecord":
			var body struct {
				Host, Type, Value string
				ZID               int
			}
			if !decode(&body) {
				break
			}
			if body.Host != name || body.ZID != 42 || body.Type != "TXT" {
				reject("incorrect record target")
				break
			}
			create(body.Value, result(map[string]string{"RecordID": "record-7"}))
		case "POST DeleteRecord":
			var body struct{ RecordID string }
			if !decode(&body) {
				break
			}
			if body.RecordID != "record-7" {
				reject("incorrect record ID")
				break
			}
			remove(result(map[string]string{}))
		default:
			reject("unexpected DNS operation")
		}
	}
	return true
}
