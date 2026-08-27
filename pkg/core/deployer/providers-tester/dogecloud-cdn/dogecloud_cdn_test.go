//go:build tester

package dogecloudcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/dogecloud-cdn"
)

var (
	fp            = tester.InitArgs("DOGECLOUDCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./dogecloud_cdn_test.go -args \
	--DOGECLOUDCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--DOGECLOUDCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--DOGECLOUDCDN_ACCESSKEY="your-access-key" \
	--DOGECLOUDCDN_SECRETKEY="your-secret-key" \
	--DOGECLOUDCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
			Domain:    fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
