//go:build tester

package aliyunmse_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-mse"
)

var (
	fp               = tester.InitArgs("ALIYUNMSE_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fServiceType     string
	fGatewayId       string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fServiceType, "SERVICETYPE")
	fp.DefineString(&fGatewayId, "GATEWAYID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aliyun_mse_test.go -args \
	--ALIYUNMSE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNMSE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNMSE_ACCESSKEYID="your-access-key-id" \
	--ALIYUNMSE_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNMSE_REGION="cn-hangzhou" \
	--ALIYUNMSE_SERVICETYPE="cloudnative" \
	--ALIYUNMSE_GATEWAYID="your-mse-gateway-id" \
	--ALIYUNMSE_DOMAIN="example.com"
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
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
