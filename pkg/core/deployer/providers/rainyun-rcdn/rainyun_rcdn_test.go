package rainyunrcdn_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/rainyun-rcdn"
)

var (
	fp            = tester.Args("RAINYUNRCDN_")
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

	go test -v ./rainyun_rcdn_test.go -args \
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
