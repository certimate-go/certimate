package k8ssecret

import (
	"strings"
	"testing"
)

const testKubeConfig = `apiVersion: v1
kind: Config
clusters:
  - name: test
    cluster:
      server: https://127.0.0.1:6443
      insecure-skip-tls-verify: true
contexts:
  - name: test
    context:
      cluster: test
      user: test
current-context: test
users:
  - name: test
    user:
      token: test-token
`

// Secrets live in the core API group, which is served under "/api".
// If APIPath is left empty, rest.RESTClientFor builds requests against
// "/v1/namespaces/..." and the API server answers 404
// ("the server could not find the requested resource").
func TestCreateK8sClientSetsCoreAPIPath(t *testing.T) {
	client, err := createK8sClient(testKubeConfig)
	if err != nil {
		t.Fatalf("createK8sClient() returned an unexpected error: %v", err)
	}

	const want = "/api/v1"
	if got := client.Get().URL().Path; !strings.HasPrefix(got, want) {
		t.Errorf("request path = %q, want it to start with %q", got, want)
	}
}
