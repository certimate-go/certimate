package tencentcloudclb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-clb"
)

var (
	fp              = tester.Args("TENCENTCLOUDCLB_")
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

	go test -v ./tencentcloud_clb_test.go -args \
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
			ResourceType:   impl.RESOURCE_TYPE_LOADBALANCER,
			LoadbalancerId: fLoadbalancerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:       fSecretId,
			SecretKey:      fSecretKey,
			Region:         fRegion,
			ResourceType:   impl.RESOURCE_TYPE_LISTENER,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})

	t.Run("Deploy_ToRuleDomain", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:       fSecretId,
			SecretKey:      fSecretKey,
			Region:         fRegion,
			ResourceType:   impl.RESOURCE_TYPE_RULEDOMAIN,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
			Domain:         fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
