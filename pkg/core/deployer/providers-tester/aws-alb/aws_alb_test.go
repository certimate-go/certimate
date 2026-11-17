//go:build tester

package awsalb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-alb"
)

var (
	fp               = tester.InitArgs("AWSALB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fLoadbalancerArn string
	fListenerArn     string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerArn, "LOADBALANCERARN")
	fp.DefineString(&fListenerArn, "LISTENERARN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./aws_alb_test.go -args \
	--AWSALB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSALB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSALB_ACCESSKEYID="your-access-key-id" \
	--AWSALB_SECRETACCESSKEY="your-secret-access-id" \
	--AWSALB_REGION="us-east-1" \
	--AWSALB_LOADBALANCERARN="your-loadbalancer-arn" \
	--AWSALB_LISTENERARN="your-listener-arn"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			SecretAccessKey:   fSecretAccessKey,
			Region:            fRegion,
			LoadbalancerArn:   fLoadbalancerArn,
			ListenerArn:       fListenerArn,
			CertificateSource: impl.CERTIFICATE_SOURCE_ACM,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
