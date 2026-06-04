package awsclb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/aws-clb"
)

var (
	fp                = tester.Args("AWSCLB_")
	fTestCertPath     string
	fTestKeyPath      string
	fAccessKeyId      string
	fSecretAccessKey  string
	fRegion           string
	fLoadbalancerName string
	fLoadbalancerPort int
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerName, "LOADBALANCERNAME")
	fp.DefineInt(&fLoadbalancerPort, "LOADBALANCERPORT")
}

/*
Shell command to run this test:

	go test -v ./aws_clb_test.go -args \
	--AWSCLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSCLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSCLB_ACCESSKEYID="your-access-key-id" \
	--AWSCLB_SECRETACCESSKEY="your-secret-access-id" \
	--AWSCLB_REGION="us-east-1" \
	--AWSCLB_LOADBALANCERNAME="your-loadbalancer-name" \
	--AWSCLB_LOADBALANCERPORT=443
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:       fAccessKeyId,
			SecretAccessKey:   fSecretAccessKey,
			Region:            fRegion,
			LoadbalancerName:  fLoadbalancerName,
			LoadbalancerPort:  int32(fLoadbalancerPort),
			CertificateSource: impl.CERTIFICATE_SOURCE_ACM,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
