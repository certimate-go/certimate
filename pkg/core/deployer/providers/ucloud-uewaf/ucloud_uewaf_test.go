package uclouduewaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uewaf"
)

var (
	fp            = tester.Args("UCLOUDUEWAF_")
	fTestCertPath string
	fTestKeyPath  string
	fPrivateKey   string
	fPublicKey    string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./ucloud_uewaf_test.go -args \
	--UCLOUDUEWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUEWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUEWAF_PRIVATEKEY="your-private-key" \
	--UCLOUDUEWAF_PUBLICKEY="your-public-key" \
	--UCLOUDUEWAF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
			Domain:     fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
