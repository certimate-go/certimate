//go:build tester

package ctcccloudcms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/ctcccloud-cms"
)

var (
	fp               = tester.InitArgs("CTCCCLOUDCMS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./ctcccloud_cms_test.go -args \
	--CTCCCLOUDCMS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDCMS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDCMS_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDCMS_SECRETACCESSKEY="your-secret-access-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
