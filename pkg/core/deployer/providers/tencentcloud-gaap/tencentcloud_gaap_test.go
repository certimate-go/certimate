package tencentcloudgaap_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-gaap"
)

var (
	fp            = tester.Args("TENCENTCLOUDCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fProxyId      string
	fListenerId   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fProxyId, "PROXYID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_gaap_test.go -args \
	--TENCENTCLOUDGAAP_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDGAAP_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDGAAP_SECRETID="your-secret-id" \
	--TENCENTCLOUDGAAP_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDGAAP_PROXYID="your-gaap-group-id" \
	--TENCENTCLOUDGAAP_LISTENERID="your-clb-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:     fSecretId,
			SecretKey:    fSecretKey,
			DeployTarget: impl.DEPLOY_TARGET_LISTENER,
			ProxyId:      fProxyId,
			ListenerId:   fListenerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
