package mohuamvh_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/mohua-mvh"
)

var (
	fp            = tester.Args("MOHUAMVH_")
	fTestCertPath string
	fTestKeyPath  string
	fUsername     string
	fApiPassword  string
	fHostID       string
	fDomainID     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fApiPassword, "APIPASSWORD")
	fp.DefineString(&fHostID, "HOSTID")
	fp.DefineString(&fDomainID, "DOMAINID")
}

/*
Shell command to run this test:

	go test -v ./mohuamvh_test.go -args \
	--MOHUAMVH_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--MOHUAMVH_TESTKEYPATH="/path/to/your-test-key.pem" \
	--MOHUAMVH_USERNAME="your-username" \
	--MOHUAMVH_APIPASSWORD="your-api-password" \
	--MOHUAMVH_HOSTID="your-virtual-host-id" \
	--MOHUAMVH_DOMAINID="your-domain-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Username:    fUsername,
			ApiPassword: fApiPassword,
			HostId:      fHostID,
			DomainId:    fDomainID,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
