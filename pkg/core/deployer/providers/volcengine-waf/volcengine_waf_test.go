package volcenginewaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-waf"
)

var (
	fp               = tester.Args("VOLCENGINEWAF_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fAccessMode      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fAccessMode, "ACCESSMODE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./volcengine_waf_test.go -args \
	--VOLCENGINEWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINEWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINEWAF_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINEWAF_ACCESSKEYSECRET="your-access-key-secret" \
	--VOLCENGINEWAF_REGION="cn-beijing" \
	--VOLCENGINEWAF_ACCESSMODE="cname" \
	--VOLCENGINEWAF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			AccessMode:      fAccessMode,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
