//go:build tester

package aliyunmse_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-mse"
)

var (
	fp                  = tester.InitArgs("ALIYUNMSE_")
	fTestCertPath       string
	fTestKeyPath        string
	fAccessKeyId        string
	fAccessKeySecret    string
	fRegion             string
	fGatewayId          string
	fDomainMatchPattern string
	fDomain             string
	fForceHttps         bool
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fGatewayId, "GATEWAYID")
	fp.DefineString(&fDomainMatchPattern, "DOMAINMATCHPATTERN")
	fp.DefineString(&fDomain, "DOMAIN")
	fp.DefineBool(&fForceHttps, "FORCEHTTPS")
}

/*
Shell command to run this test (enables HTTPS and updates the certificate for selected domains):

	go test -tags=tester -v ./aliyun_mse_test.go -args \
	--ALIYUNMSE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNMSE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNMSE_ACCESSKEYID="your-access-key-id" \
	--ALIYUNMSE_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNMSE_REGION="cn-hangzhou" \
	--ALIYUNMSE_GATEWAYID="gw-your-test-gateway" \
	--ALIYUNMSE_DOMAINMATCHPATTERN="exact" \
	--ALIYUNMSE_DOMAIN="example.com" \
	--ALIYUNMSE_FORCEHTTPS=false
*/
func TestProvider(t *testing.T) {
	fp.Parse()
	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			Region:             fRegion,
			GatewayId:          fGatewayId,
			DomainMatchPattern: fDomainMatchPattern,
			Domain:             fDomain,
			ForceHttps:         fForceHttps,
		})
		require.NoError(t, err)
		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
