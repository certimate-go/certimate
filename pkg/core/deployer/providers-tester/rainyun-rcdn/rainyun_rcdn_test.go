//go:build tester

package rainyunrcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/rainyun-rcdn"
)

var (
	fp            = tester.InitArgs("RAINYUNRCDN_")
	fTestCertPath string
	fTestKeyPath  string
	fApiKey       string
	fInstanceId   int64
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fApiKey, "APIKEY")
	fp.DefineInt64(&fInstanceId, "INSTANCEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./rainyun_rcdn_test.go -args \
	--RAINYUNRCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--RAINYUNRCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--RAINYUNRCDN_APIKEY="your-api-key" \
	--RAINYUNRCDN_INSTANCEID="your-rcdn-instance-id" \
	--RAINYUNRCDN_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			ApiKey:             fApiKey,
			InstanceId:         fInstanceId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
