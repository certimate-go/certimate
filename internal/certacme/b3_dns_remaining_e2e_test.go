package certacme_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type b3DNSProviderConfig struct {
	Access   map[string]any
	Provider string
	Apply    map[string]any
	Env      []string
}

func b3DNSConfig(t *testing.T, f *b3DNSFixture, family string) b3DNSProviderConfig {
	c := b3DNSProviderConfig{Provider: family, Access: map[string]any{"apiToken": "fixture-dns-token"}}
	switch family {
	case "linode", "digitalocean":
		c.Access = map[string]any{"accessToken": "fixture-dns-token"}
	case "qingcloud":
		c.Provider = "qingcloud-dns"
		c.Access = map[string]any{"accessKeyId": "fixture-dns-id", "secretAccessKey": "fixture-dns-secret"}
	case "bunny":
		c.Access = map[string]any{"apiKey": "fixture-dns-token"}
	case "cpanel":
		c.Access = map[string]any{"serverUrl": "https://cpanel.example.test", "username": "fixture-user", "apiToken": "fixture-dns-token"}
	case "cloudflare":
		c.Access["apiTokenForZone"] = "fixture-zone-token"
		c.Apply = map[string]any{"dnsTTL": 120}
	case "vercel":
		c.Access = map[string]any{"apiAccessToken": "fixture-dns-token", "teamId": "fixture-team"}
	case "googlecloud", "oraclecloud":
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		f.credentialKey = key
		privateKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
		if family == "googlecloud" {
			account, _ := json.Marshal(map[string]any{"type": "service_account", "project_id": "fixture-project", "private_key_id": "fixture-key", "private_key": privateKey, "client_email": "fixture@fixture-project.iam.gserviceaccount.com", "token_uri": "https://oauth2.googleapis.com/token"})
			c.Access = map[string]any{"projectId": "fixture-project", "serviceAccountKey": string(account)}
			c.Provider = "googlecloud-dns"
		} else {
			c.Access = map[string]any{"authMethod": "apikey", "privateKey": privateKey, "publicKeyFingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff", "tenancyOcid": "ocid1.tenancy.fixture", "userOcid": "ocid1.user.fixture"}
			c.Provider = "oraclecloud-dns"
			c.Apply = map[string]any{"providerConfig": map[string]any{"region": "us-phoenix-1", "compartmentOcid": "ocid1.compartment.fixture"}}
		}
	case "yandexcloud":
		return b3YandexDNSConfig(t, f)
	}
	return c
}

func (f *b3DNSFixture) remainingCloud(w http.ResponseWriter, r *http.Request) bool {
	if f.family == "linode" || f.family == "qingcloud" || f.family == "yandexcloud" {
		return false
	}
	reply := func(status int, v any) {
		w.WriteHeader(status)
		if v != nil {
			json.NewEncoder(w).Encode(v)
		}
	}
	fail := func() {
		f.t.Errorf("unhandled %s DNS %s %s", f.family, r.Method, r.URL)
		reply(404, map[string]string{"message": "unexpected request"})
	}
	auth := func(header, expected string) bool {
		if r.Header.Get(header) == expected {
			return true
		}
		f.t.Errorf("%s DNS credential mismatch on %s", f.family, r.URL)
		reply(403, map[string]string{"message": "credential mismatch"})
		return false
	}
	create := func(name, kind, value string) {
		name = strings.TrimSuffix(name, ".")
		if name != "_acme-challenge."+f.domain && name != strings.TrimSuffix("_acme-challenge."+f.domain, ".example.test") {
			f.t.Errorf("%s TXT name mismatch %q", f.family, name)
		}
		if kind != "TXT" || value != f.expected {
			f.t.Errorf("%s TXT payload mismatch type=%s value=%s expected=%s", f.family, kind, value, f.expected)
		}
		f.value = value
		f.creates++
	}
	remove := func() {
		if f.value == "" {
			f.t.Errorf("%s cleanup without record", f.family)
		}
		f.value = ""
		f.deletes++
	}
	switch f.family {
	case "bunny":
		if !auth("AccessKey", "fixture-dns-token") {
			return true
		}
		switch {
		case r.Method == "GET" && r.URL.Path == "/dnszone":
			records := []any{}
			if f.value != "" {
				records = append(records, map[string]any{"Id": 7, "Type": 3, "Name": strings.TrimSuffix("_acme-challenge."+f.domain, ".example.test"), "Value": f.value})
			}
			reply(200, map[string]any{"Items": []any{map[string]any{"Id": 42, "Domain": "example.test", "Records": records}}, "CurrentPage": 1, "TotalItems": 1, "HasMoreItems": false})
		case r.Method == "PUT" && r.URL.Path == "/dnszone/42/records":
			var body struct {
				Name, Value string
				Type        int
			}
			json.NewDecoder(r.Body).Decode(&body)
			if body.Type != 3 {
				f.t.Error("Bunny record type mismatch")
			}
			create(body.Name, "TXT", body.Value)
			reply(201, map[string]any{"Id": 7})
		case r.Method == "DELETE" && r.URL.Path == "/dnszone/42/records/7":
			remove()
			reply(204, nil)
		default:
			fail()
		}
	case "cloudflare":
		expectedToken := "Bearer fixture-dns-token"
		if r.URL.Path == "/client/v4/zones" {
			expectedToken = "Bearer fixture-zone-token"
		}
		if !auth("Authorization", expectedToken) {
			return true
		}
		result := func(v any) {
			reply(200, map[string]any{"success": true, "errors": []any{}, "result": v, "result_info": map[string]int{"page": 1, "total_pages": 1, "count": 1, "total_count": 1}})
		}
		switch {
		case r.Method == "GET" && r.URL.Path == "/client/v4/zones":
			result([]any{map[string]any{"id": "fixture-zone", "name": "example.test", "status": "active"}})
		case r.Method == "POST" && r.URL.Path == "/client/v4/zones/fixture-zone/dns_records":
			var body struct{ Name, Type, Content string }
			json.NewDecoder(r.Body).Decode(&body)
			if body.Content != fmt.Sprintf("%q", f.expected) {
				f.t.Error("Cloudflare quoted TXT content mismatch")
			}
			create(body.Name, body.Type, strings.Trim(body.Content, "\""))
			result(map[string]string{"id": "fixture-record"})
		case r.Method == "DELETE" && r.URL.Path == "/client/v4/zones/fixture-zone/dns_records/fixture-record":
			remove()
			result(map[string]string{"id": "fixture-record"})
		default:
			fail()
		}
	case "digitalocean", "vercel", "netlify":
		if !auth("Authorization", "Bearer fixture-dns-token") {
			return true
		}
		if f.family == "vercel" && r.URL.Query().Get("teamId") != "fixture-team" {
			f.t.Error("Vercel team mismatch")
		}
		path := "/v2/domains/example.test/records"
		if f.family == "netlify" {
			path = "/api/v1/dns_zones/example_test/dns_records"
		}
		switch {
		case r.Method == "POST" && r.URL.Path == path:
			var body struct{ Name, Hostname, Type, Value, Data string }
			json.NewDecoder(r.Body).Decode(&body)
			name, value := body.Name, body.Value
			if f.family == "netlify" {
				name = body.Hostname
			}
			if f.family == "digitalocean" {
				value = body.Data
			}
			create(name, body.Type, value)
			response := map[string]any{"id": "7"}
			if f.family == "digitalocean" {
				response = map[string]any{"domain_record": map[string]int{"id": 7}}
			}
			if f.family == "vercel" {
				response = map[string]any{"uid": "7"}
			}
			reply(201, response)
		case r.Method == "DELETE" && r.URL.Path == path+"/7":
			remove()
			reply(204, nil)
		default:
			fail()
		}
	case "gcore":
		if !auth("Authorization", "APIKey fixture-dns-token") {
			return true
		}
		path := "/dns/v2/zones/example.test/" + "_acme-challenge." + f.domain + "/TXT"
		switch {
		case r.Method == "GET" && r.URL.Path == "/dns/v2/zones/example.test":
			reply(200, map[string]string{"name": "example.test"})
		case r.Method == "GET" && r.URL.Path == path:
			if f.value == "" {
				reply(404, map[string]string{"error": "not found"})
			} else {
				reply(200, map[string]any{"ttl": 60, "resource_records": []any{map[string]any{"content": []string{f.value}}}})
			}
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/dns/v2/zones/"):
			reply(404, map[string]string{"error": "not found"})
		case (r.Method == "POST" || r.Method == "PUT") && r.URL.Path == path:
			var body struct {
				Records []struct {
					Content []string `json:"content"`
				} `json:"resource_records"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			if len(body.Records) != 1 || len(body.Records[0].Content) != 1 {
				f.t.Error("Gcore record shape mismatch")
				reply(400, nil)
				return true
			}
			create("_acme-challenge."+f.domain, "TXT", body.Records[0].Content[0])
			reply(200, map[string]any{})
		case r.Method == "DELETE" && r.URL.Path == path:
			remove()
			reply(204, nil)
		default:
			fail()
		}
	case "cpanel":
		if !auth("Authorization", "cpanel fixture-user:fixture-dns-token") {
			return true
		}
		if r.URL.Query().Get("zone") != "example.test" {
			f.t.Error("cPanel zone mismatch")
		}
		b64 := func(v string) string { return base64.StdEncoding.EncodeToString([]byte(v)) }
		switch r.URL.Path {
		case "/execute/DNS/parse_zone":
			records := []any{map[string]any{"type": "record", "record_type": "SOA", "dname_b64": b64("example.test."), "data_b64": []string{b64("ns.example.test."), b64("admin.example.test."), b64("1")}}}
			if f.value != "" {
				records = append(records, map[string]any{"type": "record", "record_type": "TXT", "line_index": 7, "dname_b64": b64("_acme-challenge." + f.domain + "."), "data_b64": []string{b64(f.value)}})
			}
			reply(200, map[string]any{"status": 1, "data": records})
		case "/execute/DNS/mass_edit_zone":
			if raw := r.URL.Query().Get("add"); raw != "" {
				var body struct {
					Name string   `json:"dname"`
					Type string   `json:"record_type"`
					Data []string `json:"data"`
				}
				json.Unmarshal([]byte(raw), &body)
				if len(body.Data) != 1 {
					f.t.Error("cPanel record shape mismatch")
					reply(400, nil)
					return true
				}
				create(body.Name, body.Type, body.Data[0])
			} else if r.URL.Query().Get("remove") == "7" {
				remove()
			} else {
				fail()
				return true
			}
			reply(200, map[string]any{"status": 1, "data": map[string]string{"new_serial": "2"}})
		default:
			fail()
		}
	case "googlecloud":
		if r.Host == "oauth2.googleapis.com" {
			r.ParseForm()
			parts := strings.Split(r.Form.Get("assertion"), ".")
			if len(parts) != 3 {
				f.t.Error("Google JWT missing")
				reply(400, nil)
				return true
			}
			signature, _ := base64.RawURLEncoding.DecodeString(parts[2])
			hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
			if rsa.VerifyPKCS1v15(&f.credentialKey.PublicKey, crypto.SHA256, hash[:], signature) != nil {
				f.t.Error("Google JWT signature mismatch")
			}
			reply(200, map[string]any{"access_token": "fixture-dns-token", "token_type": "Bearer", "expires_in": 3600})
			return true
		}
		if !auth("Authorization", "Bearer fixture-dns-token") {
			return true
		}
		prefix := "/dns/v1/projects/fixture-project/managedZones"
		record := func() map[string]any {
			return map[string]any{"name": "_acme-challenge." + f.domain + ".", "type": "TXT", "ttl": 60, "rrdatas": []string{fmt.Sprintf("%q", f.value)}}
		}
		switch {
		case r.Method == "GET" && r.URL.Path == prefix:
			reply(200, map[string]any{"managedZones": []any{map[string]string{"name": "fixture-zone", "dnsName": "example.test.", "visibility": "public"}}})
		case r.Method == "GET" && r.URL.Path == prefix+"/fixture-zone/rrsets":
			records := []any{}
			if f.value != "" {
				records = append(records, record())
			}
			reply(200, map[string]any{"rrsets": records})
		case r.Method == "POST" && r.URL.Path == prefix+"/fixture-zone/changes":
			var body struct {
				Additions, Deletions []struct {
					Name, Type string
					Data       []string `json:"rrdatas"`
				}
			}
			json.NewDecoder(r.Body).Decode(&body)
			if len(body.Additions) == 1 && len(body.Additions[0].Data) == 1 {
				record := body.Additions[0]
				create(record.Name, record.Type, strings.Trim(record.Data[0], "\""))
			} else if len(body.Deletions) == 1 {
				remove()
			} else {
				f.t.Error("Google change shape mismatch")
			}
			reply(200, map[string]string{"id": "7", "status": "done"})
		default:
			fail()
		}
	case "oraclecloud":
		if !f.verifyOracleSignature(r) {
			f.t.Error("Oracle DNS signature mismatch")
			reply(403, nil)
			return true
		}
		path := "/20180115/zones/example.test./records/" + "_acme-challenge." + f.domain
		if r.URL.Path != path {
			fail()
			return true
		}
		switch r.Method {
		case "PATCH":
			var body struct {
				Items []struct{ Domain, Rdata, Rtype, Operation, RecordHash string }
			}
			json.NewDecoder(r.Body).Decode(&body)
			if len(body.Items) != 1 {
				f.t.Error("Oracle record shape mismatch")
				reply(400, nil)
				return true
			}
			item := body.Items[0]
			if item.Operation == "REMOVE" && item.RecordHash == "fixture-hash" {
				remove()
			} else {
				create(item.Domain, item.Rtype, item.Rdata)
			}
			reply(200, map[string]any{"items": []any{}})
		case "GET":
			w.Header().Set("opc-total-items", "1")
			reply(200, map[string]any{"items": []any{map[string]string{"domain": "_acme-challenge." + f.domain, "rtype": "TXT", "rdata": fmt.Sprintf("%q", f.value), "recordHash": "fixture-hash"}}})
		default:
			fail()
		}
	default:
		return false
	}
	return true
}

func (f *b3DNSFixture) verifyOracleSignature(r *http.Request) bool {
	fields := map[string]string{}
	for _, part := range strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Signature "), ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok {
			fields[key] = strings.Trim(value, "\"")
		}
	}
	if fields["keyId"] != "ocid1.tenancy.fixture/ocid1.user.fixture/00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff" {
		return false
	}
	var lines []string
	for _, header := range strings.Fields(fields["headers"]) {
		value := r.Header.Get(header)
		if header == "(request-target)" {
			value = strings.ToLower(r.Method) + " " + r.URL.RequestURI()
		}
		if header == "host" {
			value = r.Host
		}
		if header == "content-length" {
			value = fmt.Sprint(r.ContentLength)
		}
		lines = append(lines, header+": "+value)
	}
	signature, err := base64.StdEncoding.DecodeString(fields["signature"])
	if err != nil {
		return false
	}
	hash := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return rsa.VerifyPKCS1v15(&f.credentialKey.PublicKey, crypto.SHA256, hash[:], signature) == nil
}
