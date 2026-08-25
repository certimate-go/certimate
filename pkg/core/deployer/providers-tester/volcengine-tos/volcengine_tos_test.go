//go:build tester

package volcenginetos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/volcengine-tos"
)

var (
	fp               = tester.InitArgs("VOLCENGINETOS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
	fBucket          string
	fDomain          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./volcengine_tos_test.go -args \
	--VOLCENGINETOS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINETOS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINETOS_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINETOS_SECRETACCESSKEY="your-secret-access-key" \
	--VOLCENGINETOS_REGION="cn-beijing" \
	--VOLCENGINETOS_BUCKET="your-tos-bucket" \
	--VOLCENGINETOS_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
			Bucket:          fBucket,
			Domain:          fDomain,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
