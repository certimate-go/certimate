package kong_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/kong"
)

var (
	fp             = tester.Args("KONG_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiToken      string
	fCertificateId string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./kong_test.go -args \
	--KONG_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--KONG_TESTKEYPATH="/path/to/your-test-key.pem" \
	--KONG_SERVERURL="http://127.0.0.1:9080" \
	--KONG_APITOKEN="your-admin-token" \
	--KONG_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiToken:                 fApiToken,
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
