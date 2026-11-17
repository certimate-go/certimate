//go:build tester

package tencentcloudddospro_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ddospro"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDDDOSPRO_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fInstanceId   string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fInstanceId, "INSTANCEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_ddospro_test.go -args \
	--TENCENTCLOUDDDOSPRO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDDDOSPRO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDDDOSPRO_SECRETID="your-secret-id" \
	--TENCENTCLOUDDDOSPRO_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDDDOSPRO_INSTANCEID="your-instance-id" \
	--TENCENTCLOUDDDOSPRO_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			InstanceId:         fInstanceId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
