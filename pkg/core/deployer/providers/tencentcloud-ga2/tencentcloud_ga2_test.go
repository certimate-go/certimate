package tencentcloudga2_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/tencentcloud-ga2"
)

var (
	fp             = tester.Args("TENCENTCLOUDGA2_")
	fTestCertPath  string
	fTestKeyPath   string
	fSecretId      string
	fSecretKey     string
	fAcceleratorId string
	fListenerId    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSecretId, "SECRETID")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fAcceleratorId, "ACCELERATORID")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -v ./tencentcloud_ga2_test.go -args \
	--TENCENTCLOUDGA2_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--TENCENTCLOUDGA2_TESTKEYPATH="/path/to/your-test-key.pem" \
	--TENCENTCLOUDGA2_SECRETID="your-secret-id" \
	--TENCENTCLOUDGA2_SECRETKEY="your-secret-key" \
	--TENCENTCLOUDGA2_ACCELERATORID="your-ga2-accelerator-id" \
	--TENCENTCLOUDGA2_LISTENERID="your-ga2-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			SecretId:      fSecretId,
			SecretKey:     fSecretKey,
			DeployTarget:  impl.DEPLOY_TARGET_LISTENER,
			AcceleratorId: fAcceleratorId,
			ListenerId:    fListenerId,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
