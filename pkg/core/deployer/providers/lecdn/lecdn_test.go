package lecdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/lecdn"
)

var (
	fp             = tester.Args("LECDN_")
	fTestCertPath  string
	fTestKeyPath   string
	fServerUrl     string
	fApiVersion    string
	fUsername      string
	fPassword      string
	fCertificateId int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fApiVersion, "APIVERSION", "v3")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineInt64(&fCertificateId, "CERTIFICATEID")
}

/*
Shell command to run this test:

	go test -v ./lecdn_test.go -args \
	--LECDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--LECDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--LECDN_SERVERURL="http://127.0.0.1:5090" \
	--LECDN_USERNAME="your-username" \
	--LECDN_PASSWORD="your-password" \
	--LECDN_CERTIFICATEID="your-certificate-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToCertificate", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiVersion:               fApiVersion,
			ApiRole:                  "client",
			Username:                 fUsername,
			Password:                 fPassword,
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
