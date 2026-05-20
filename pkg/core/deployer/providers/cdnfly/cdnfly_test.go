package cdnfly_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/cdnfly"
)

var (
	fp             = tester.Args("CDNFLY_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiKey        string
	fApiSecret     string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fApiSecret, "APISECRET")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./cdnfly_test.go -args \
	--CDNFLY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CDNFLY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CDNFLY_SERVERURL="http://127.0.0.1:88" \
	--CDNFLY_APIKEY="your-api-key" \
	--CDNFLY_APISECRET="your-api-secret" \
	--CDNFLY_CERTIFICATEID="your-cert-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			ApiSecret:                fApiSecret,
			AllowInsecureConnections: true,
			DeployTarget:             impl.DEPLOY_TARGET_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
