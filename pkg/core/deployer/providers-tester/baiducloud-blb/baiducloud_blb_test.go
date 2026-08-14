//go:build tester

package baiducloudblb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/baiducloud-blb"
)

var (
	fp               = tester.InitArgs("BAIDUCLOUDBLB_")
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

	go test -tags=tester -v ./baiducloud_blb_test.go -args \
	--BAIDUCLOUDBLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAIDUCLOUDBLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAIDUCLOUDBLB_ACCESSKEYID="your-access-key-id" \
	--BAIDUCLOUDBLB_SECRETACCESSKEY="your-secret-access-key" \
	--BAIDUCLOUDBLB_REGION="bj" \
	--BAIDUCLOUDBLB_LOADBALANCERID="your-blb-loadbalancer-id" \
	--BAIDUCLOUDBLB_DOMAIN="your-blb-sni-domain"
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
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
