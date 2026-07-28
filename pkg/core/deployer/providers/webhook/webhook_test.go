package webhook_test

import (
	"testing"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/webhook"
	tester "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp                  = tester.Args("WEBHOOK_")
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

	go test -v ./webhook_test.go -args \
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
