//go:build tester

package tencentcloudcss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-css"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDCSS_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_css_test.go -args \
	--TENCENTCLOUDCSS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDCSS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDCSS_SECRETID="your-secret-id" \
	--TENCENTCLOUDCSS_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDCSS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:  fSecretId,
			SecretKey: fSecretKey,
			Domain:    fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
