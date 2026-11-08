package axisnow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListCertificatesSubscribedTotalCompatibility(t *testing.T) {
	tests := []struct {
		name  string
		total string
		want  string
	}{
		{name: "numeric_zero", total: `0`, want: "0"},
		{name: "quoted_zero", total: `"0"`, want: "0"},
		{name: "numeric_count", total: `12`, want: "12"},
		{name: "quoted_count", total: `"12"`, want: "12"},
		{name: "null", total: `null`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"success":true,"result_info":{"page":1,"per_page":100,"count":1,"total_count":1,"subscribed_total":` + tt.total + `},"result":[{"uuid":"existing-cert","name":"example.com"}]}`
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(body))
			}))
			defer srv.Close()
			c, err := NewClient(WithApiToken("test-only"))
			if err != nil {
				t.Fatal(err)
			}
			c.rc.SetBaseURL(srv.URL)
			res, err := c.ListCertificates(&ListCertificatesRequest{})
			if err != nil {
				t.Fatal(err)
			}
			if res.ResultInfo.TotalCount != 1 || len(res.Results) != 1 || res.Results[0].UUID != "existing-cert" {
				t.Fatalf("certificate pagination or identity lost: %#v", res.ResultInfo)
			}
			if got := res.ResultInfo.SubscribedTotal.String(); got != tt.want {
				t.Errorf("subscribed_total = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResponseErrorsDoNotIncludeCertificatePayload(t *testing.T) {
	secret := "PRIVATE-CERTIFICATE-PAYLOAD"
	for _, tt := range []struct {
		name   string
		status int
	}{
		{name: "invalid_json", status: http.StatusOK},
		{name: "http_error", status: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(strings.Repeat(secret, 10000)))
			}))
			defer srv.Close()
			c, err := NewClient(WithApiToken("test-only"))
			if err != nil {
				t.Fatal(err)
			}
			c.rc.SetBaseURL(srv.URL)
			_, err = c.ListCertificates(&ListCertificatesRequest{})
			if err == nil || len(err.Error()) > 1000 || strings.Contains(err.Error(), secret) {
				t.Fatalf("response payload leaked or error not bounded: %v", err)
			}
		})
	}
}

func TestInvalidSubscribedTotalIsRejected(t *testing.T) {
	var response ListCertificatesResponse
	if err := json.Unmarshal([]byte(`{"result_info":{"subscribed_total":"bad"}}`), &response); err == nil {
		t.Fatal("expected malformed numeric value to be rejected")
	}
}
