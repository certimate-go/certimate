package safeline_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/safeline"
)

var (
	fp             = tester.Args("SAFELINE_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiToken      string
	fCertificateId int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineInt64(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./safeline_test.go -args \
	--SAFELINE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--SAFELINE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--SAFELINE_SERVERURL="http://127.0.0.1:9443" \
	--SAFELINE_APITOKEN="your-api-token" \
	--SAFELINE_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiToken:                 fApiToken,
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
