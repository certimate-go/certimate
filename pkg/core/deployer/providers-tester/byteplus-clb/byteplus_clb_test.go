//go:build tester

package byteplusclb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/byteplus-clb"
)

var (
	fp               = tester.InitArgs("BYTEPLUSCLB_")
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

	go test -tags=tester -v ./byteplus_clb_test.go -args \
	--BYTEPLUSCLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BYTEPLUSCLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BYTEPLUSCLB_ACCESSKEYID="your-access-key-id" \
	--BYTEPLUSCLB_SECRETACCESSKEY="your-secret-access-key" \
	--BYTEPLUSCLB_REGION="ap-southeast-1" \
	--BYTEPLUSCLB_LISTENERID="your-listener-id"
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

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
