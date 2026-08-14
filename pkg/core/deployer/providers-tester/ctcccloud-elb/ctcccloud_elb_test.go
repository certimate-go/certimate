//go:build tester

package ctcccloudelb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ctcccloud-elb"
)

var (
	fp               = tester.InitArgs("CTCCCLOUDELB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegionId        string
	fLoadbalancerId  string
	fListenerId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegionId, "REGIONID")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./ctcccloud_elb_test.go -args \
	--CTCCCLOUDELB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDELB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDELB_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDELB_SECRETACCESSKEY="your-secret-access-key" \
	--CTCCCLOUDELB_REGIONID="your-region-id" \
	--CTCCCLOUDELB_LOADBALANCERID="your-elb-instance-id" \
	--CTCCCLOUDELB_LISTENERID="your-elb-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToLoadbalancer", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
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
			SecretAccessKey: fSecretAccessKey,
			RegionId:        fRegionId,
			DeployTarget:    impl.DEPLOY_TARGET_LISTENER,
			ListenerId:      fListenerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
