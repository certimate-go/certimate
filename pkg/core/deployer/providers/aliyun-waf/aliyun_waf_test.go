package aliyunwaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-waf"
)

var (
	fp               = tester.Args("ALIYUNWAF_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fInstanceId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fInstanceId, "INSTANCEID")
}

/*
Shell command to run this test:

	go test -v ./aliyun_waf_test.go -args \
	--ALIYUNWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNWAF_ACCESSKEYID="your-access-key-id" \
	--ALIYUNWAF_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNWAF_REGION="cn-hangzhou" \
	--ALIYUNWAF_INSTANCEID="your-waf-instance-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
			InstanceId:      fInstanceId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
