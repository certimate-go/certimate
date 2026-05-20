package uclouduclb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uclb"
)

var (
	fp              = tester.Args("UCLOUDUCLB_")
	fTestCertPath   string
	fTestKeyPath    string
	fPrivateKey     string
	fPublicKey      string
	fRegion         string
	fLoadbalancerId string
	fVServerId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fVServerId, "VSERVERID")
}

/*
Shell command to run this test:

	go test -v ./ucloud_uclb_test.go -args \
	--UCLOUDUCLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUCLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUCLB_PRIVATEKEY="your-private-key" \
	--UCLOUDUCLB_PUBLICKEY="your-public-key" \
	--UCLOUDUCLB_REGION="cn-bj2" \
	--UCLOUDUCLB_LOADBALANCERID="your-loadbalancer-id" \
	--UCLOUDUCLB_VSERVERID="your-vserver-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToVServer", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey:     fPrivateKey,
			PublicKey:      fPublicKey,
			Region:         fRegion,
			DeployTarget:   impl.DEPLOY_TARGET_VSERVER,
			LoadbalancerId: fLoadbalancerId,
			VServerId:      fVServerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
