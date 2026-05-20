package ucloudupathx_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-upathx"
)

var (
	fp             = tester.Args("UCLOUDUPATHX_")
	fTestCertPath  string
	fTestKeyPath   string
	fPrivateKey    string
	fPublicKey     string
	fRegion        string
	fAcceleratorId string
	fListenerPort  int
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPublicKey, "PUBLICKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fAcceleratorId, "ACCELERATORID")
	fp.DefineInt(&fListenerPort, "LISTENERPORT", 443)
}

/*
Shell command to run this test:

	go test -v ./ucloud_upathx_test.go -args \
	--UCLOUDUPATHX_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--UCLOUDUPATHX_TESTKEYPATH="/path/to/your-test-key.pem" \
	--UCLOUDUPATHX_PRIVATEKEY="your-private-key" \
	--UCLOUDUPATHX_PUBLICKEY="your-public-key" \
	--UCLOUDUPATHX_ACCELERATORID="your-uga-id" \
	--UCLOUDUPATHX_ACCELERATORPORT="443"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			PrivateKey:    fPrivateKey,
			PublicKey:     fPublicKey,
			AcceleratorId: fAcceleratorId,
			ListenerPort:  int32(fListenerPort),
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
