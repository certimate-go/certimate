package volcengineimagex_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-imagex"
)

var (
	fp               = tester.Args("VOLCENGINEIMAGEX_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fServiceId       string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fServiceId, "SERVICEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./volcengine_imagex_test.go -args \
	--VOLCENGINEIMAGEX_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEIMAGEX_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEIMAGEX_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEIMAGEX_ACCESSKEYSECRET="your-access-key-secret" \
	--VOLCENGINEIMAGEX_REGION="cn-north-1" \
	--VOLCENGINEIMAGEX_SERVICEID="your-service-id" \
	--VOLCENGINEIMAGEX_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			ServiceId:       fServiceId,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
