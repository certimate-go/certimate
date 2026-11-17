//go:build tester

package tencentcloudvod_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-vod"
)

var (
	fp            = tester.InitArgs("TENCENTCLOUDVOD_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fDomain       string
	fSubAppId     int64
	fInstanceId   string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineInt64(&fSubAppId, "SUBAPPID")
	fp.DefineString(&fInstanceId, "INSTANCEID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_vod_test.go -args \
	--TENCENTCLOUDVOD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDVOD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDVOD_SECRETID="your-secret-id" \
	--TENCENTCLOUDVOD_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDVOD_SUBAPPID="your-app-id" \
	--TENCENTCLOUDVOD_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:           fSecretId,
			SecretKey:          fSecretKey,
			SubAppId:           fSubAppId,
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
