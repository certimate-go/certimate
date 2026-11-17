package certacme

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDNSPersistAuthorizationSemantics(t *testing.T) {
	account := "https://ca.example/acct/Case"
	expected := dnsPersistValue{issuer: "pki.goog", accountURI: account, wildcard: true}
	base := "pki.goog; accounturi=" + account
	for _, tc := range []struct {
		name     string
		value    string
		wildcard bool
		ready    bool
	}{
		{name: "field order whitespace and issuer case", value: " PKI.GOOG ; POLICY = WILDCARD ; AccountURI = " + account + "\t", wildcard: true, ready: true},
		{name: "exact account URI", value: strings.Replace(base, "/Case", "/case", 1) + "; policy=wildcard", wildcard: true},
		{name: "different issuer", value: "other.example; accounturi=" + account + "; policy=wildcard", wildcard: true},
		{name: "wildcard requires policy", value: base, wildcard: true},
		{name: "wildcard permits apex", value: base + "; policy=wildcard", ready: true},
		{name: "unknown policy exact name", value: base + "; policy=unknown", ready: true},
		{name: "empty policy exact name", value: base + "; policy=", ready: true},
		{name: "unknown policy excludes wildcard", value: base + "; policy=unknown", wildcard: true},
		{name: "duplicate account", value: base + "; ACCOUNTURI=" + account},
		{name: "duplicate policy", value: base + "; policy=wildcard; POLICY=wildcard", wildcard: true},
		{name: "duplicate unknown tag", value: base + "; future=a; FUTURE=b"},
		{name: "unknown parameter", value: base + "; future=ok; policy=wildcard", wildcard: true, ready: true},
		{name: "not yet expired", value: base + "; persistUntil=1700000001", ready: true},
		{name: "expires now", value: base + "; persistUntil=1700000000"},
		{name: "expired", value: base + "; persistUntil=1699999999"},
		{name: "invalid expiration", value: base + "; persistUntil=tomorrow"},
		{name: "invalid whitespace", value: base + "; future=two words"},
		{name: "trailing delimiter", value: base + ";"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected.wildcard = tc.wildcard
			if got := matchesDNSPersistValue(tc.value, expected, time.Unix(1700000000, 0), nil, nil); got != tc.ready {
				t.Fatalf("record readiness = %v, want %v", got, tc.ready)
			}
		})
	}
	t.Run("offered issuer and original account URI", func(t *testing.T) {
		expected := dnsPersistValue{issuer: "first.example", accountURI: "https://ca.example/alias/Case", wildcard: true}
		value := base + "; policy=wildcard"
		if !matchesDNSPersistValue(value, expected, time.Now(), []string{"first.example", "PKI.GOOG"}, []string{account}) {
			t.Fatal("an offered alternate issuer and the exact registered account URI must be accepted")
		}
		if matchesDNSPersistValue(strings.Replace(value, "/Case", "/case", 1), expected, time.Now(), []string{"pki.goog"}, []string{account}) {
			t.Fatal("account URI aliases must remain case-sensitive")
		}
	})
}

type dnsPersistRoundTripFunc func(*http.Request) (*http.Response, error)

func (f dnsPersistRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDNSPersistHTTPContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var requestContext context.Context
	transport := &dnsPersistTransport{ctx: ctx, transport: dnsPersistRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestContext = req.Context()
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("directory"))}, nil
	})}
	req, err := http.NewRequest(http.MethodGet, "https://ca.example/directory", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if err := requestContext.Err(); err != nil {
		t.Fatalf("request canceled before its response body was read: %v", err)
	}
	cancel()
	select {
	case <-requestContext.Done():
	case <-time.After(time.Second):
		t.Fatal("the task cancellation did not reach the directory request")
	}
}
