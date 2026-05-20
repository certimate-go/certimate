package aliyunesa_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-esa"
)

var (
	fp               = tester.Args("ALIYUNESA_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fSiteId          int64
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineInt64(&fSiteId, "SITEID")
}

/*
Shell command to run this test:

	go test -v ./aliyun_esa_test.go -args \
	--ALIYUNESA_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNESA_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNESA_ACCESSKEYID="your-access-key-id" \
	--ALIYUNESA_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNESA_REGION="cn-hangzhou" \
	--ALIYUNESA_SITEID="your-esa-site-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			SiteId:          fSiteId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
