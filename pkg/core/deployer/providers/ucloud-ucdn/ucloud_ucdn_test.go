package uclouducdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-ucdn"
)

var (
	fp            = tester.Args("UCLOUDUCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fPrivateKey   string
	fPublicKey    string
	fDomainId     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fDomainId, "DOMAINID")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ucdn_test.go -args \
	--UCLOUDUCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUCDN_PRIVATEKEY="your-private-key" \
	--UCLOUDUCDN_PUBLICKEY="your-public-key" \
	--UCLOUDUCDN_DOMAINID="your-domain-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
			DomainId:   fDomainId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
