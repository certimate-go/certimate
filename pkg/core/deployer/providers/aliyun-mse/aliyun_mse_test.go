package aliyunmse

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	alimse "github.com/certimate-go/certimate/pkg/sdk3rd-trimmed/github.com/alibabacloud-go/mse-20190531/v5/client"
)

func newTestDeployer(t *testing.T, handler http.HandlerFunc) *Deployer {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	d, err := NewDeployer(&DeployerConfig{
		AccessKeyId: "test-key", AccessKeySecret: "test-secret",
		Region: "cn-hangzhou", GatewayId: "gw-test", Domain: "example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "mse.cn-hangzhou.aliyuncs.com", tea.StringValue(d.sdkClient.Endpoint))
	d.sdkClient.Endpoint = tea.String(strings.TrimPrefix(server.URL, "http://"))
	d.sdkClient.Protocol = tea.String("HTTP")
	d.SetLogger(slog.New(slog.DiscardHandler))
	return d
}

func TestUpdateDomainCertificate(t *testing.T) {
	testCases := []struct {
		name        string
		protocol    string
		certificate string
		action      string
		request     url.Values
		response    string
	}{
		{
			name: "enable HTTPS without redirect", protocol: "HTTP",
			action: "UpdateGatewayDomain",
			request: url.Values{
				"GatewayUniqueId": {"gw-test"}, "Id": {"42"}, "CertIdentifier": {"new-cn-hangzhou"},
				"Protocol": {"HTTPS"}, "MustHttps": {"false"},
			},
			response: `{"Success":true,"Data":42}`,
		},
		{
			name: "replace HTTPS certificate only", protocol: "HTTPS", certificate: "old-cn-hangzhou",
			action: "UpdateSSLCert",
			request: url.Values{
				"GatewayUniqueId": {"gw-test"}, "DomainId": {"42"}, "CertIdentifier": {"new-cn-hangzhou"},
			},
			response: `{"Success":true,"Data":true}`,
		},
		{
			name: "skip unchanged certificate", protocol: "HTTPS", certificate: "new-cn-hangzhou",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			d := newTestDeployer(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				assert.Equal(t, tc.action, r.Header.Get("x-acs-action"))
				assert.NoError(t, r.ParseForm())
				assert.Equal(t, tc.request, r.Form)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.response))
			})
			err := d.updateDomainCertificate(t.Context(), &alimse.ListGatewayDomainResponseBodyData{
				Id: tea.Int64(42), Name: tea.String("example.com"), Protocol: tea.String(tc.protocol),
				CertIdentifier: tea.String(tc.certificate), MustHttps: tea.Bool(true),
			}, "new-cn-hangzhou")
			require.NoError(t, err)
			if tc.action == "" {
				assert.Zero(t, calls.Load())
			} else {
				assert.EqualValues(t, 1, calls.Load())
			}
		})
	}
}

func TestUpdateDomainCertificateErrors(t *testing.T) {
	testCases := []struct {
		name     string
		protocol string
		status   int
		response string
	}{
		{"HTTPS API error", "HTTPS", http.StatusForbidden, `{"Code":"NoPermission","Message":"denied"}`},
		{"HTTPS request rejected", "HTTPS", http.StatusOK, `{"Success":false,"Data":true}`},
		{"HTTPS certificate not updated", "HTTPS", http.StatusOK, `{"Success":true,"Data":false}`},
		{"HTTP request rejected", "HTTP", http.StatusOK, `{"Success":false,"Data":42}`},
		{"HTTP wrong domain result", "HTTP", http.StatusOK, `{"Success":true,"Data":99}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newTestDeployer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.response))
			})
			err := d.updateDomainCertificate(t.Context(), &alimse.ListGatewayDomainResponseBodyData{
				Id: tea.Int64(42), Protocol: tea.String(tc.protocol), CertIdentifier: tea.String("old-cn-hangzhou"),
			}, "new-cn-hangzhou")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "mse.Update")
		})
	}
}
