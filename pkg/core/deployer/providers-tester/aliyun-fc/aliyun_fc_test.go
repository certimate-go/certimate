//go:build tester

package aliyunfc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aliyun-fc"
)

var (
	fp               = tester.InitArgs("ALIYUNFC_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aliyun_fc_test.go -args \
	--ALIYUNFC_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNFC_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNFC_ACCESSKEYID="your-access-key-id" \
	--ALIYUNFC_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNFC_REGION="cn-hangzhou" \
	--ALIYUNFC_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:        fAccessKeyId,
			AccessKeySecret:    fAccessKeySecret,
			Region:             fRegion,
			ServiceVersion:     "3.0",
			DomainMatchPattern: impl.DOMAIN_MATCH_PATTERN_EXACT,
			Domain:             fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
