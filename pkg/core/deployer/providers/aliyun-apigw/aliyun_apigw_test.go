package aliyunapigw_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-apigw"
)

var (
	fp               = tester.Args("ALIYUNAPIGW_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fServiceType     string
	fGatewayId       string
	fGroupId         string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fGatewayId, "GATEWAYID")
	fp.DefineString(&fGroupId, "GROUPID")
	fp.DefineString(&fServiceType, "SERVICETYPE")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./aliyun_apigw_test.go -args \
	--ALIYUNAPIGW_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNAPIGW_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNAPIGW_ACCESSKEYID="your-access-key-id" \
	--ALIYUNAPIGW_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNAPIGW_REGION="cn-hangzhou" \
	--ALIYUNAPIGW_SERVICETYPE="cloudnative" \
	--ALIYUNAPIGW_GATEWAYID="your-api-gateway-id" \
	--ALIYUNAPIGW_GROUPID="your-api-group-id" \
	--ALIYUNAPIGW_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			Region:             fRegion,
			ServiceType:        fServiceType,
			GatewayId:          fGatewayId,
			GroupId:            fGroupId,
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
