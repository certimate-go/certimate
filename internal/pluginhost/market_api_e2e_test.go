package pluginhost_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/pluginhost"
)

type marketAPI struct {
	t     *testing.T
	base  string
	token string
}

type marketResponse struct {
	Code int
	Msg  string
	Data json.RawMessage
}

func (a marketAPI) request(method, path string, body any, result any) int {
	a.t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		a.t.Fatal(err)
	}
	req, err := http.NewRequest(method, a.base+path, bytes.NewReader(raw))
	if err != nil {
		a.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Set("Authorization", a.token)
	}
	client := http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer response.Body.Close()
	if result != nil {
		if err := json.NewDecoder(response.Body).Decode(result); err != nil {
			a.t.Fatalf("%s %s: %v", method, path, err)
		}
	}
	return response.StatusCode
}

func (a marketAPI) market(method, path string, body any, wantCode int, reason string, result any) {
	a.t.Helper()
	var response marketResponse
	status := a.request(method, "/api/plugin/market"+path, body, &response)
	if status != 200 || response.Code != wantCode {
		a.t.Fatalf("%s %s: HTTP %d, code %d (%s), want %d", method, path, status, response.Code, response.Msg, wantCode)
	}
	if reason != "" && !bytes.Contains(response.Data, []byte(`"reason":"`+reason+`"`)) {
		a.t.Fatalf("%s %s: wanted %s, got %s", method, path, reason, response.Data)
	}
	if result != nil {
		if err := json.Unmarshal(response.Data, result); err != nil {
			a.t.Fatalf("%s: %v", path, err)
		}
	}
}

func (a marketAPI) login(email string) marketAPI {
	a.t.Helper()
	collection := "_superusers"
	if email != fixtureEmail {
		collection = "users"
	}
	var result struct{ Token string }
	if status := a.request("POST", "/api/collections/"+collection+"/auth-with-password", map[string]string{"identity": email, "password": fixturePassword}, &result); status != 200 || result.Token == "" {
		a.t.Fatalf("login %s: HTTP %d", email, status)
	}
	a.token = result.Token
	return a
}

func startMarketFixture(t *testing.T, state string) (marketAPI, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	path := filepath.Join(t.TempDir(), "server.log")
	logs, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer logs.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMarketBrowserFixture$", "-test.v", "-test.timeout=2m")
	cmd.Env = append(os.Environ(), "CERTIMATE_TEST_MARKET_ADDRESS="+address, "CERTIMATE_TEST_MARKET_STATE="+state)
	cmd.Stdout, cmd.Stderr = logs, logs
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	stop := func() { cancel(); <-done }
	t.Cleanup(func() {
		stop()
		if t.Failed() {
			data, _ := os.ReadFile(path)
			t.Logf("fixture server log:\n%s", data)
		}
	})
	base := "http://" + address
	client := http.Client{Timeout: time.Second}
	for deadline := time.Now().Add(35 * time.Second); time.Now().Before(deadline); {
		response, err := client.Get(base + "/api/health")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == 200 {
				return marketAPI{t: t, base: base}, stop
			}
		}
		select {
		case <-done:
			t.Fatal("fixture exited during setup")
		case <-time.After(50 * time.Millisecond):
		}
	}
	t.Fatal("fixture did not start")
	return marketAPI{}, stop
}

func waitMarketInstall(t *testing.T, api marketAPI, provider, source string, want pluginhost.JobState) {
	t.Helper()
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); {
		var job pluginhost.JobStatus
		api.market("GET", "/install/status/"+provider, nil, 0, "", &job)
		if job.Source.ID != source {
			t.Fatalf("job source: %+v", job)
		}
		if job.State.Terminal() {
			if job.State != want {
				t.Fatalf("job result: %+v", job)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("installation did not finish")
}

func TestE2E_CustomMarketsHTTP(t *testing.T) {
	state := t.TempDir()
	anonymous, stop := startMarketFixture(t, state)
	api := anonymous.login(fixtureEmail)
	regular := anonymous.login("regular@example.test")
	add := func(api marketAPI, repo string) pluginhost.MarketSourceStatus {
		t.Helper()
		var status pluginhost.MarketSourceStatus
		api.market("POST", "/sources", map[string]string{"repository": "https://github.com/" + repo}, 0, "", &status)
		return status
	}
	for _, client := range []marketAPI{anonymous, regular} {
		for _, path := range []string{"/sources", "/install"} {
			if status := client.request("POST", "/api/plugin/market"+path, map[string]string{}, nil); status != 401 && status != 403 {
				t.Fatalf("unauthorized %s returned %d", path, status)
			}
		}
	}
	api.market("POST", "/sources", map[string]string{"repository": "https://github.com/fixture-missing/plugins"}, 502, "", nil)
	if status := api.request("POST", "/fixture/catalog/invalid", nil, nil); status != 204 {
		t.Fatalf("set invalid catalog: HTTP %d", status)
	}
	api.market("POST", "/sources", map[string]string{"repository": "https://github.com/" + fixtureRepoA}, 502, "", nil)
	var sources []pluginhost.MarketSourceStatus
	api.market("GET", "/sources", nil, 0, "", &sources)
	if len(sources) != 1 || sources[0].ID != "official" {
		t.Fatalf("failed additions left configured sources: %+v", sources)
	}
	stop()
	anonymous, stop = startMarketFixture(t, state)
	api = anonymous.login(fixtureEmail)
	api.market("GET", "/sources", nil, 0, "", &sources)
	if len(sources) != 1 || sources[0].ID != "official" {
		t.Fatalf("failed additions were persisted: %+v", sources)
	}
	a, b := add(api, fixtureRepoA), add(api, fixtureRepoB)
	api.market("POST", "/sources", map[string]string{"repository": "https://github.com/FIXTURE-A/Plugins.git/"}, 409, "duplicate_source", nil)
	if _, err := os.Stat(filepath.Join(state, "plugins", fixtureProvider)); !os.IsNotExist(err) {
		t.Fatalf("adding a market installed a binary: %v", err)
	}

	page := func(query string) pluginhost.MarketPage {
		t.Helper()
		var result pluginhost.MarketPage
		api.market("GET", query, nil, 0, "", &result)
		return result
	}
	card := func(list pluginhost.MarketPage, source, provider string) pluginhost.MarketCard {
		t.Helper()
		for _, card := range list.Items {
			if card.ProviderType == provider && card.Source.ID == source {
				return card
			}
		}
		t.Fatalf("missing card for %s/%s: %+v", source, provider, list.Items)
		return pluginhost.MarketCard{}
	}
	list := page("")
	if !card(list, "official", "fixture-legacy").CanUpdate || !card(list, "official", "fixture-legacy").Official ||
		card(list, "", "fixture-manual").Official || card(list, "official", "fixture-manual").CanUpdate ||
		card(list, a.ID, fixtureProvider).Official || !card(list, b.ID, fixtureProvider).CanInstall {
		t.Fatalf("bad source identity or eligibility: %+v", list.Items)
	}
	api.market("POST", "/install", map[string]string{"providerType": fixtureConsumer, "source": b.ID}, 0, "", nil)
	waitMarketInstall(t, api, fixtureConsumer, b.ID, pluginhost.JobFailed)
	api.market("POST", "/install", map[string]string{"providerType": fixtureProvider, "source": a.ID}, 0, "", nil)
	api.market("POST", "/install", map[string]string{"providerType": fixtureProvider, "source": a.ID}, 409, "install_in_progress", nil)
	api.market("POST", "/install", map[string]string{"providerType": fixtureProvider, "source": b.ID}, 409, "ownership_conflict", nil)
	api.market("DELETE", "/sources/"+a.ID, nil, 409, "source_busy", nil)
	waitMarketInstall(t, api, fixtureProvider, a.ID, pluginhost.JobInstalled)
	api.market("POST", "/install", map[string]string{"providerType": fixtureConsumer, "source": b.ID}, 0, "", nil)
	waitMarketInstall(t, api, fixtureConsumer, b.ID, pluginhost.JobInstalled)
	api.market("POST", "/install", map[string]string{"providerType": fixtureDuplicateOwner, "source": b.ID}, 0, "", nil)
	waitMarketInstall(t, api, fixtureDuplicateOwner, b.ID, pluginhost.JobFailed)
	api.market("POST", "/update/"+fixtureProvider, map[string]string{"source": b.ID}, 409, "ownership_conflict", nil)
	api.market("DELETE", "/"+fixtureProvider+"?source="+b.ID, nil, 409, "ownership_conflict", nil)
	if conflict := card(page(""), b.ID, fixtureProvider); conflict.CanInstall || conflict.CanUpdate || conflict.CanDelete {
		t.Fatalf("conflicting card has actions: %+v", conflict)
	}

	observed := make(chan string, 12)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var result struct{ Version, Message string }
		_ = json.NewDecoder(r.Body).Decode(&result)
		observed <- fmt.Sprintf("%s:%s:%s", r.Header.Get("Authorization"), result.Version, result.Message)
	}))
	defer upstream.Close()
	var access struct{ ID string }
	if status := api.request("POST", "/api/collections/access/records", map[string]any{
		"name": "Fixture access", "provider": fixtureAccess, "reserve": "notif", "config": map[string]string{"url": upstream.URL, "token": "fixture-token"},
	}, &access); status != 200 || access.ID == "" {
		t.Fatalf("save access: HTTP %d", status)
	}
	notify := func(provider, version string) {
		t.Helper()
		var result marketResponse
		status := api.request("POST", "/api/notifications/test", map[string]string{"provider": provider, "accessId": access.ID}, &result)
		if status != 200 || result.Code != 0 {
			t.Fatalf("notify: HTTP %d %+v", status, result)
		}
		select {
		case got := <-observed:
			if got != "Bearer fixture-token:"+version+":Welcome to use Certimate!" {
				t.Fatalf("unexpected notification: %s", got)
			}
		default:
			t.Fatal("notification did not reach target")
		}
	}
	workflowNotify := func() {
		t.Helper()
		node := func(id, kind string, config map[string]any) map[string]any {
			return map[string]any{"id": id, "type": kind, "data": map[string]any{"name": id, "config": config}}
		}
		graph := map[string]any{"nodes": []any{
			node("start", "start", nil),
			node("notify", "bizNotify", map[string]any{"provider": fixtureConsumer, "providerAccessId": access.ID, "subject": "Market workflow", "message": "Shared credential"}),
			node("end", "end", nil),
		}}
		var workflow struct{ ID string }
		if status := api.request("POST", "/api/collections/workflow/records", map[string]any{
			"name": "Market consumer workflow", "trigger": "manual", "enabled": true,
			"hasContent": true, "hasDraft": false, "graphContent": graph, "graphDraft": graph,
		}, &workflow); status != 200 || workflow.ID == "" {
			t.Fatalf("save workflow: HTTP %d", status)
		}
		var started struct {
			Code int
			Msg  string
			Data struct{ RunID string }
		}
		if status := api.request("POST", "/api/workflows/"+workflow.ID+"/runs", map[string]string{"trigger": "manual"}, &started); status != 200 || started.Code != 0 || started.Data.RunID == "" {
			t.Fatalf("start workflow: HTTP %d %+v", status, started)
		}
		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
			var run struct{ Status, Error string }
			if status := api.request("GET", "/api/collections/workflow_run/records/"+started.Data.RunID, nil, &run); status != 200 {
				t.Fatalf("get workflow run: HTTP %d", status)
			}
			switch run.Status {
			case "succeeded":
				select {
				case got := <-observed:
					if got != "Bearer fixture-token:1.0.0:Shared credential" {
						t.Fatalf("workflow delivery: %s", got)
					}
				default:
					t.Fatal("workflow notification did not arrive")
				}
				return
			case "failed", "canceled":
				t.Fatalf("workflow did not succeed: %+v", run)
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatal("workflow did not complete")
	}
	notify(fixtureProvider, "1.0.0")
	workflowNotify()
	stop()
	anonymous, _ = startMarketFixture(t, state)
	api = anonymous.login(fixtureEmail)
	if restored := card(page(""), a.ID, fixtureProvider); !restored.CanDelete {
		t.Fatalf("restart lost ownership: %+v", restored)
	}
	api.market("DELETE", "/sources/"+a.ID, nil, 0, "", nil)
	if retained := card(page(""), a.ID, fixtureProvider); !retained.CanDelete || retained.SourceAvailable || retained.CanUpdate {
		t.Fatalf("wrong retained card: %+v", retained)
	}
	workflowNotify()
	notify(fixtureProvider, "1.0.0")
	api.market("POST", "/update/"+fixtureProvider, nil, 404, "", nil)
	api.market("DELETE", "/"+fixtureProvider+"?source="+a.ID, nil, 409, "access_owner_in_use", nil)
	if readded := add(api, fixtureRepoA); readded.ID != a.ID {
		t.Fatalf("re-add changed ID: %+v", readded)
	}

	for _, condition := range []string{"invalid", "broken", "valid"} {
		if status := api.request("POST", "/fixture/catalog/"+condition, nil, nil); status != 204 {
			t.Fatalf("fixture state: HTTP %d", status)
		}
		var refreshed pluginhost.MarketSourceStatus
		api.market("POST", "/sources/"+a.ID+"/refresh", nil, 0, "", &refreshed)
		switch condition {
		case "invalid":
			if !refreshed.Stale || refreshed.Error == "" || card(page(""), b.ID, fixtureConsumer).Status != "installed" {
				t.Fatalf("catalog failure was not isolated: %+v", refreshed)
			}
		case "broken":
			receipt := filepath.Join(state, "plugins", fixtureProvider, ".market.json")
			before, _ := os.ReadFile(receipt)
			api.market("POST", "/update/"+fixtureProvider, map[string]string{"source": a.ID}, 500, "", nil)
			after, _ := os.ReadFile(receipt)
			if !bytes.Equal(before, after) || len(after) == 0 {
				t.Fatal("failed update changed receipt")
			}
			notify(fixtureProvider, "1.0.0")
		case "valid":
			api.market("POST", "/update/"+fixtureProvider, map[string]string{"source": a.ID}, 0, "", nil)
			notify(fixtureProvider, "2.0.0")
		}
	}
	api.market("POST", "/update/fixture-legacy", nil, 0, "", nil)
	api.market("POST", "/update/fixture-manual", nil, 409, "ownership_conflict", nil)
	api.market("DELETE", "/"+fixtureConsumer+"?source="+b.ID, nil, 0, "", nil)
	api.market("DELETE", "/sources/"+a.ID, nil, 0, "", nil)
	api.market("DELETE", "/"+fixtureProvider+"?source="+a.ID, nil, 0, "", nil)
	if _, err := os.Stat(filepath.Join(state, "plugins", fixtureProvider)); !os.IsNotExist(err) {
		t.Fatalf("plugin files remain: %v", err)
	}
	orphan := filepath.Join(state, "plugins", ".dl-"+fixtureProvider+".zip")
	if err := os.WriteFile(orphan, []byte("partial bytes from a different asset"), 0o600); err != nil {
		t.Fatal(err)
	}
	api.market("POST", "/install", map[string]string{"providerType": fixtureProvider, "source": b.ID}, 0, "", nil)
	waitMarketInstall(t, api, fixtureProvider, b.ID, pluginhost.JobInstalled)
	api.market("DELETE", "/"+fixtureProvider+"?source="+a.ID, nil, 409, "ownership_conflict", nil)
	notify(fixtureProvider, "9.0.0")
	api.market("DELETE", "/"+fixtureProvider, nil, 0, "", nil)
}

func TestE2E_IndependentPluginMarket(t *testing.T) {
	if os.Getenv("CERTIMATE_TEST_INDEPENDENT_CATALOG") == "" {
		t.Skip("prepare the independent project catalog per docs/verification/custom-plugin-markets.md")
	}
	anonymous, _ := startMarketFixture(t, t.TempDir())
	api := anonymous.login(fixtureEmail)
	var market pluginhost.MarketSourceStatus
	api.market("POST", "/sources", map[string]string{"repository": "https://github.com/independent-fixture/catalog"}, 0, "", &market)
	if market.Error != "" || market.Official {
		t.Fatalf("independent market: %+v", market)
	}
	api.market("POST", "/install", map[string]string{"providerType": "mocknotify", "source": market.ID}, 0, "", nil)
	waitMarketInstall(t, api, "mocknotify", market.ID, pluginhost.JobInstalled)
	deliveries := make(chan string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data struct{ Message string }
		_ = json.NewDecoder(r.Body).Decode(&data)
		deliveries <- data.Message
	}))
	defer target.Close()
	var access struct{ ID string }
	if status := api.request("POST", "/api/collections/access/records", map[string]any{
		"name": "Independent webhook", "provider": "webhook", "reserve": "notif", "config": map[string]string{"url": target.URL, "allowInsecureConnections": "false"},
	}, &access); status != 200 || access.ID == "" {
		t.Fatalf("save access: HTTP %d", status)
	}
	var result marketResponse
	if status := api.request("POST", "/api/notifications/test", map[string]string{"provider": "mocknotify", "accessId": access.ID}, &result); status != 200 || result.Code != 0 {
		t.Fatalf("notification: HTTP %d %+v", status, result)
	}
	select {
	case got := <-deliveries:
		if got != "Welcome to use Certimate!" {
			t.Fatalf("unexpected delivery %q", got)
		}
	default:
		t.Fatal("market-installed independent plugin did not deliver")
	}
}
