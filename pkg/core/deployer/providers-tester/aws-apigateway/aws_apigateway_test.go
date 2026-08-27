//go:build tester

package awsapigateway_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-apigateway"
)

var (
	fp               = tester.InitArgs("AWSAPIGATEWAY_")
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

	go test -tags=tester -v ./aws_apigateway_test.go -args \
	--AWSAPIGATEWAY_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSAPIGATEWAY_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSAPIGATEWAY_ACCESSKEYID="your-access-key-id" \
	--AWSAPIGATEWAY_SECRETACCESSKEY="your-secret-access-id" \
	--AWSAPIGATEWAY_REGION="us-east-1" \
	--AWSAPIGATEWAY_DOMAIN="example.com"
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
