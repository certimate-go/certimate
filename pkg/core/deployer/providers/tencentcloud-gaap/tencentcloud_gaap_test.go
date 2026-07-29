package tencentcloudgaap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-gaap"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("TENCENTCLOUGAAP_")
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
	--TENCENTCLOUDGAAP_LISTENERID="your-gaap-listener-id"
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
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
