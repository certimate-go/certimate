//go:build tester

package qiniucdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/qiniu-cdn"
)

var (
	fp            = tester.InitArgs("QINIUCDN_")
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

	go test -tags=tester -v ./qiniu_cdn_test.go -args \
	--QINIUCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINIUCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINIUCDN_ACCESSKEY="your-access-key" \
	--QINIUCDN_SECRETKEY="your-secret-key" \
	--QINIUCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKey:          fAccessKey,
			SecretKey:          fSecretKey,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
