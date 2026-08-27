//go:build tester

package qingcloudlb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/qingcloud-lb"
)

var (
	fp               = tester.InitArgs("QINGCLOUDLB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fZoneId          string
	fCertificateId   string
	fLoadbalancerId  string
	fListenerId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fZoneId, "ZONEID")
	fp.DefineString(&fLoadbalancerId, "LOADBALANCERID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./qingcloud_lb_test.go -args \
	--QINGCLOUDLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINGCLOUDLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINGCLOUDLB_ACCESSKEYID="your-access-key-id" \
	--QINGCLOUDLB_SECRETACCESSKEY="your-secret-access-key" \
	--QINGCLOUDLB_ZONEID="pek3a" \
	--QINGCLOUDLB_LOADBALANCERID="your-lb-loadbalancer-id" \
	--QINGCLOUDLB_LISTENERID="your-lb-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToLoadbalancer", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			ZoneId:          fZoneId,
			DeployTarget:    impl.DEPLOY_TARGET_LOADBALANCER,
			LoadbalancerId:  fLoadbalancerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToListenerId", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			ZoneId:          fZoneId,
			DeployTarget:    impl.DEPLOY_TARGET_LISTENER,
			ListenerId:      fListenerId,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
