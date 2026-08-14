//go:build tester

package tencentcloudgaap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-gaap"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUGAAP_")
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

	go test -tags=tester -v ./tencentcloud_gaap_test.go -args \
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

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
