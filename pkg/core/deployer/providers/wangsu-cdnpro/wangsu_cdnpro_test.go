package wangsucdnpro_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/wangsu-cdnpro"
)

var (
	fp               = tester.Args("WANGSUCDNPRO_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fApiKey          string
	fEnvironment     string
	fDomain          string
	fCertificateId   string
	fWebhookId       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fEnvironment, "ENVIRONMENT", "production")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
	fp.DefineString(&fWebhookId, "WEBHOOKID")
}

/*
Shell command to run this test:

	go test -v ./wangsu_cdnpro_test.go -args \
	--WANGSUCDNPRO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--WANGSUCDNPRO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--WANGSUCDNPRO_ACCESSKEYID="your-access-key-id" \
	--WANGSUCDNPRO_ACCESSKEYSECRET="your-access-key-secret" \
	--WANGSUCDNPRO_APIKEY="your-api-key" \
	--WANGSUCDNPRO_ENVIRONMENT="production" \
	--WANGSUCDNPRO_DOMAIN="example.com" \
	--WANGSUCDNPRO_CERTIFICATEID="your-certificate-id" \
	--WANGSUCDNPRO_WEBHOOKID="your-webhook-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			ApiKey:          fApiKey,
			Environment:     fEnvironment,
			Domain:          fDomain,
			CertificateId:   fCertificateId,
			WebhookId:       fWebhookId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
