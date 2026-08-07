package jdcloudwaf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-waf"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp               = it.Args("JDCLOUDWAF_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fInstanceId      string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fInstanceId, "INSTANCEID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./jdcloud_waf_test.go -args \
	--JDCLOUDWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--JDCLOUDWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--JDCLOUDWAF_ACCESSKEYID="your-access-key-id" \
	--JDCLOUDWAF_ACCESSKEYSECRET="your-secret-access-key" \
	--JDCLOUDWAF_INSTANCEID="your-instance-id" \
	--JDCLOUDWAF_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			InstanceId:      fInstanceId,
			Domain:          fDomain,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
