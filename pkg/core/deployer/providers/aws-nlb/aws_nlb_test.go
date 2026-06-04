package awsnlb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-nlb"
)

var (
	fp               = tester.Args("AWSNLB_")
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

	go test -v ./aws_nlb_test.go -args \
	--AWSNLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSNLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSNLB_ACCESSKEYID="your-access-key-id" \
	--AWSNLB_SECRETACCESSKEY="your-secret-access-id" \
	--AWSNLB_REGION="us-east-1" \
	--AWSNLB_LOADBALANCERARN="your-loadbalancer-arn" \
	--AWSNLB_LISTENERARN="your-listener-arn"
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
