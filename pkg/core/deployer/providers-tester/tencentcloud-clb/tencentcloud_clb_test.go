//go:build tester

package tencentcloudclb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-clb"
)

var (
	fp              = tester.InitArgs("TENCENTCLOUDCLB_")
	fTestCertPath   string
	fTestKeyPath    string
	fSecretId       string
	fSecretKey      string
	fRegion         string
	fLoadbalancerId string
	fListenerId     string
	fDomain         string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fListenerId, "LISTENERID")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./tencentcloud_clb_test.go -args \
	--TENCENTCLOUDCLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDCLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDCLB_SECRETID="your-secret-id" \
	--TENCENTCLOUDCLB_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDCLB_REGION="ap-guangzhou" \
	--TENCENTCLOUDCLB_LOADBALANCERID="your-clb-lb-id" \
	--TENCENTCLOUDCLB_LISTENERID="your-clb-lbl-id" \
	--TENCENTCLOUDCLB_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToLoadbalancer", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:       fSecretId,
			SecretKey:      fSecretKey,
			Region:         fRegion,
			DeployTarget:   impl.DEPLOY_TARGET_LOADBALANCER,
			LoadbalancerId: fLoadbalancerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:       fSecretId,
			SecretKey:      fSecretKey,
			Region:         fRegion,
			DeployTarget:   impl.DEPLOY_TARGET_LISTENER,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToRuleDomain", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:       fSecretId,
			SecretKey:      fSecretKey,
			Region:         fRegion,
			DeployTarget:   impl.DEPLOY_TARGET_RULEDOMAIN,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
			Domain:         fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
