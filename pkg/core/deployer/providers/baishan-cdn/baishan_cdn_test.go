package baishancdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baishan-cdn"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("BAISHANCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fApiToken     string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./baishan_cdn_test.go -args \
	--BAISHANCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAISHANCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAISHANCDN_APITOKEN="your-api-token" \
	--BAISHANCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiToken:           fApiToken,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
