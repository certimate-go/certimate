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
	fHostId       string
	fDomainId     int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fApiPassword, "APIPASSWORD")
	fp.DefineString(&fHostId, "HOSTID")
	fp.DefineInt64(&fDomainId, "DOMAINID")
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
			HostId:      fHostId,
			DomainId:    fDomainId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
