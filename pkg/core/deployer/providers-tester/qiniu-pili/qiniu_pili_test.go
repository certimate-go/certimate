//go:build tester

package qiniupili_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/qiniu-pili"
)

var (
	fp            = tester.InitArgs("QINIUPILI_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
	fHub          string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fHub, "HUB")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./qiniu_pili_test.go -args \
	--QINIUPILI_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINIUPILI_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINIUPILI_ACCESSKEY="your-access-key" \
	--QINIUPILI_SECRETKEY="your-secret-key" \
	--QINIUPILI_HUB="your-hub-name" \
	--QINIUPILI_DOMAIN="example.com"
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
