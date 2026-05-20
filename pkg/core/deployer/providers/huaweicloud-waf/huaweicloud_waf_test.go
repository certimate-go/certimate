package huaweicloudwaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/huaweicloud-waf"
)

var (
	fp               = tester.Args("HUAWEICLOUDWAF_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fResourceType    string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fResourceType, "RESOURCETYPE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./huaweicloud_waf_test.go -args \
	--HUAWEICLOUDWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDWAF_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDWAF_SECRETACCESSKEY="your-secret-access-key" \
	--HUAWEICLOUDWAF_REGION="cn-north-1" \
	--HUAWEICLOUDWAF_RESOURCETYPE="premium" \
	--HUAWEICLOUDWAF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			ResourceType:    fResourceType,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
