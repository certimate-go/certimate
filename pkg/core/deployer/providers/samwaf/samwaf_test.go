package samwaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/samwaf"
)

var (
	fp             = tester.Args("SAMWAF_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiKey        string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./samwaf_test.go -args \
	--SAMWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--SAMWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--SAMWAF_SERVERURL="http://127.0.0.1:26666" \
	--SAMWAF_APIKEY="your-api-key" \
	--SAMWAF_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			ResourceType:             impl.RESOURCE_TYPE_CERTIFICATE,
			CertificateId:            fCertificateId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
