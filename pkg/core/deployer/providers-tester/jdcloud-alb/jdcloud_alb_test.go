//go:build tester

package jdcloudalb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/jdcloud-alb"
)

var (
	fp               = tester.InitArgs("JDCLOUDALB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegionId        string
	fLoadbalancerId  string
	fListenerId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegionId, "REGIONID")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./jdcloud_alb_test.go -args \
	--JDCLOUDALB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--JDCLOUDALB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--JDCLOUDALB_ACCESSKEYID="your-access-key-id" \
	--JDCLOUDALB_ACCESSKEYSECRET="your-secret-access-key" \
	--JDCLOUDALB_REGION_ID="cn-north-1" \
	--JDCLOUDALB_LOADBALANCERID="your-alb-loadbalancer-id" \
	--JDCLOUDALB_LISTENERID="your-alb-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToLoadbalancer", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			RegionId:        fRegionId,
			DeployTarget:    impl.DEPLOY_TARGET_LOADBALANCER,
			LoadbalancerId:  fLoadbalancerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			RegionId:        fRegionId,
			DeployTarget:    impl.DEPLOY_TARGET_LISTENER,
			ListenerId:      fListenerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
