package tencentcloudwaf_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-waf"
)

var (
	fp            = tester.Args("TENCENTCLOUDWAF_")
	fTestCertPath string
	fTestKeyPath  string
	fSecretId     string
	fSecretKey    string
	fRegion       string
	fInstanceId   string
	fDomain       string
	fDomainId     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fInstanceId, "INSTANCEID")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineString(&fDomainId, "DOMAINID")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_waf_test.go -args \
	--TENCENTCLOUDWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDWAF_SECRETID="your-secret-id" \
	--TENCENTCLOUDWAF_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDWAF_REGION="ap-guangzhou" \
	--TENCENTCLOUDWAF_INSTANCEID="your-instance-id" \
	--TENCENTCLOUDWAF_DOMAIN="example.com" \
	--TENCENTCLOUDWAF_DOMAINID="your-domain-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:   fSecretId,
			SecretKey:  fSecretKey,
			Region:     fRegion,
			InstanceId: fInstanceId,
			Domain:     fDomain,
			DomainId:   fDomainId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
