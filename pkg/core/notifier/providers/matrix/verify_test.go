package matrix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

func TestResolveClientBaseURLWithDetail_unreachableHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := resty.New().SetTimeout(3 * time.Second)
	_, _, err := resolveClientBaseURLWithDetail(ctx, client, "http://localhost1111")
	if err == nil {
		t.Fatal("expected error for invalid host")
	}
}

func TestResolveClientBaseURLWithDetail_reachableSynapse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_matrix/client/versions":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"versions":["v1.1"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	client := resty.New().SetTimeout(5 * time.Second)
	base, _, err := resolveClientBaseURLWithDetail(ctx, client, srv.URL)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if base != srv.URL {
		t.Fatalf("base=%q want %q", base, srv.URL)
	}
}

func TestVerifyConnection_badHomeserver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := VerifyConnection(ctx, &NotifierConfig{
		HomeserverUrl: "http://localhost1111",
		AuthMode:      "token",
		AccessToken:   "dummy",
	})
	if err != nil {
		t.Fatalf("VerifyConnection: %v", err)
	}
	if res.Ok {
		t.Fatalf("expected ok=false, steps=%+v", res.Steps)
	}
	if len(res.Steps) == 0 || res.Steps[0].Ok {
		t.Fatalf("homeserver step should fail: %+v", res.Steps)
	}
}
