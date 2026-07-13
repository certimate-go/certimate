package huaweicloudvod_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-vod"
)

var (
	fp               = tester.Args("HUAWEICLOUDVOD_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./huaweicloud_vod_test.go -args \
	--HUAWEICLOUDVOD_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDVOD_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDVOD_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDVOD_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDVOD_REGION="cn-north-1" \
	--HUAWEICLOUDVOD_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			SecretAccessKey:    fSecretAccessKey,
			Region:             fRegion,
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
