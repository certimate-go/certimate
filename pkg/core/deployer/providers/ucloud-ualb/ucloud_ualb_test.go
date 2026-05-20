package ucloudualb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-ualb"
)

var (
	fp              = tester.Args("UCLOUDUALB_")
	fTestCertPath   string
	fTestKeyPath    string
	fPrivateKey     string
	fPublicKey      string
	fRegion         string
	fLoadbalancerId string
	fListenerId     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ualb_test.go -args \
	--UCLOUDUALB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUALB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUALB_PRIVATEKEY="your-private-key" \
	--UCLOUDUALB_PUBLICKEY="your-public-key" \
	--UCLOUDUALB_REGION="cn-bj2" \
	--UCLOUDUALB_LOADBALANCERID="your-loadbalancer-id" \
	--UCLOUDUALB_LISTENERID="your-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey:     fPrivateKey,
			PublicKey:      fPublicKey,
			Region:         fRegion,
			ResourceType:   impl.RESOURCE_TYPE_LISTENER,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
