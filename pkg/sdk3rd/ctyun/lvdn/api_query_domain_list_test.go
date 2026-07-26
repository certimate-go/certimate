package lvdn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/lo"

	common "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/zz-shared-common"
)

func TestDomainAPIsUseCurrentPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/domain/query-domain-list":
			if r.Method != http.MethodGet {
				t.Errorf("unexpected domain list method: %s", r.Method)
			}
			if got := r.URL.Query().Get("product_code"); got != "005" {
				t.Errorf("unexpected product_code: %s", got)
			}
			_, _ = w.Write([]byte(`{"code":100000,"message":"success","result":[]}`))
		case "/live/domain/query-domain-detail":
			if r.Method != http.MethodPost {
				t.Errorf("unexpected domain detail method: %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"code":100000,"message":"success"}`))
		case "/live/domain/update-domain":
			if r.Method != http.MethodPost {
				t.Errorf("unexpected domain update method: %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"code":100000,"message":"success"}`))
		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := common.NewClient(
		server.URL,
		common.WithAkSk("access-key-id", "secret-access-key"),
	)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	lvdnClient := &Client{domainClient: client}
	resp, err := lvdnClient.QueryDomainList(&QueryDomainListRequest{
		ProductCode: lo.ToPtr("005"),
	})
	if err != nil {
		t.Fatalf("QueryDomainList returned an error: %v", err)
	}
	if resp.ReturnObj == nil {
		t.Fatal("QueryDomainList did not normalize the response")
	}

	_, err = lvdnClient.QueryDomainDetail(&QueryDomainDetailRequest{
		Domain:      lo.ToPtr("example.com"),
		ProductCode: lo.ToPtr("005"),
	})
	if err != nil {
		t.Fatalf("QueryDomainDetail returned an error: %v", err)
	}

	_, err = lvdnClient.UpdateDomain(&UpdateDomainRequest{
		Domain:      lo.ToPtr("example.com"),
		ProductCode: lo.ToPtr("005"),
		HttpsSwitch: lo.ToPtr(int32(1)),
		CertName:    lo.ToPtr("certificate-name"),
	})
	if err != nil {
		t.Fatalf("UpdateDomain returned an error: %v", err)
	}
}
