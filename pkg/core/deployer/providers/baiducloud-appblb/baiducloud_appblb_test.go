package baiducloudappblb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baiducloud-appblb"
)

var (
	fp               = tester.Args("BAIDUCLOUDAPPBLB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fLoadbalancerId  string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./baiducloud_appblb_test.go -args \
	--BAIDUCLOUDAPPBLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAIDUCLOUDAPPBLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAIDUCLOUDAPPBLB_ACCESSKEYID="your-access-key-id" \
	--BAIDUCLOUDAPPBLB_SECRETACCESSKEY="your-secret-access-key" \
	--BAIDUCLOUDAPPBLB_REGION="bj" \
	--BAIDUCLOUDAPPBLB_LOADBALANCERID="your-blb-loadbalancer-id" \
	--BAIDUCLOUDAPPBLB_DOMAIN="your-blb-sni-domain"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			DeployTarget:    impl.DEPLOY_TARGET_LOADBALANCER,
			Region:          fRegion,
			LoadbalancerId:  fLoadbalancerId,
			Domain:          fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
