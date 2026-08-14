//go:build tester

package awsamplify_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-amplify"
)

var (
	fp               = tester.InitArgs("AWSAMPLIFY_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fAppId           string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fAppId, "APPID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aws_amplify_test.go -args \
	--AWSAMPLIFY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSAMPLIFY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSAMPLIFY_ACCESSKEYID="your-access-key-id" \
	--AWSAMPLIFY_SECRETACCESSKEY="your-secret-access-id" \
	--AWSAMPLIFY_REGION="us-east-1" \
	--AWSAMPLIFY_APPID="your-app-id" \
	--AWSAMPLIFY_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			SecretAccessKey:   fSecretAccessKey,
			Region:            fRegion,
			Domain:            fDomain,
			CertificateSource: impl.CERTIFICATE_SOURCE_ACM,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
