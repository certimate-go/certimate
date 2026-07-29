package qiniukodo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/qiniu-kodo"
	it "github.com/certimate-go/certimate/pkg/core/deployer/testing"
)

var (
	fp            = it.Args("QINIUKODO_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
	fBucket       string
	fDomain       string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fDomain, "DOMAIN")
}

/*
Shell command to run this test:

	go test -v ./qiniu_kodo_test.go -args \
	--QINIUKODO_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINIUKODO_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINIUKODO_ACCESSKEY="your-access-key" \
	--QINIUKODO_SECRETKEY="your-secret-key" \
	--QINIUKODO_BUCKET="your-bucket" \
	--QINIUKODO_DOMAIN="example.com"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
			Bucket:    fBucket,
			Domain:    fDomain,
		})
		require.NoError(t, err)

		it.TestDeploy(t, provider, it.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
