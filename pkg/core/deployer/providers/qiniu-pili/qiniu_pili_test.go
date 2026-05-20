package qiniupili_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/qiniu-pili"
)

var (
	fp            = tester.Args("QINIUPILI_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
	fHub          string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fHub, "HUB")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./qiniu_pili_test.go -args \
	--QINIUPILI_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINIUPILI_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINIUPILI_ACCESSKEY="your-access-key" \
	--QINIUPILI_SECRETKEY="your-secret-key" \
	--QINIUPILI_HUB="your-hub-name" \
	--QINIUPILI_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKey:          fAccessKey,
			SecretKey:          fSecretKey,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
