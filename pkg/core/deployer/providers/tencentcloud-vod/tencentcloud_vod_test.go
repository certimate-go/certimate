package tencentcloudvod_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-vod"
)

var (
	fp            = tester.Args("TENCENTCLOUDVOD_")
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

	go test -v ./tencentcloud_vod_test.go -args \
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
