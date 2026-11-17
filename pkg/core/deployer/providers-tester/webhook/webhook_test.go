//go:build tester

package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/webhook"
)

var (
	fp                  = tester.InitArgs("WEBHOOK_")
	fTestCertPath       string
	fTestKeyPath        string
	fWebhookUrl         string
	fWebhookContentType string
	fWebhookData        string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fWebhookUrl, "URL")
	fp.DefineString(&fWebhookContentType, "CONTENTTYPE", "application/json")
	fp.DefineString(&fWebhookData, "DATA")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./webhook_test.go -args \
	--WEBHOOK_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--WEBHOOK_TESTKEYPATH="/path/to/your-test-key.pem" \
	--WEBHOOK_URL="https://example.com/your-webhook-url" \
	--WEBHOOK_CONTENTTYPE="application/json" \
	--WEBHOOK_DATA="{\"certificate\":\"${CERTIFICATE}\",\"privateKey\":\"${PRIVATE_KEY}\"}"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			WebhookUrl:  fWebhookUrl,
			WebhookData: fWebhookData,
			Method:      "POST",
			Headers: map[string]string{
				"Content-Type": fWebhookContentType,
			},
			AllowInsecureConnections: true,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
