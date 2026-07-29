package byteplusalb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/byteplus-alb"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp               = it.Args("BYTEPLUSALB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fListenerId      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fListenerId, "LISTENERID")
}

/*
Shell command to run this test:

	go test -v ./byteplus_alb_test.go -args \
	--BYTEPLUSALB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BYTEPLUSALB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BYTEPLUSALB_ACCESSKEYID="your-access-key-id" \
	--BYTEPLUSALB_SECRETACCESSKEY="your-secret-access-key" \
	--BYTEPLUSALB_REGION="ap-southeast-1" \
	--BYTEPLUSALB_LISTENERID="your-listener-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_ToListener", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			DeployTarget:    impl.DEPLOY_TARGET_LISTENER,
			ListenerId:      fListenerId,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
