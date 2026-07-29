package qiniusslcert_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/qiniu-sslcert"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp            = it.Args("QINIUSSLCERT_")
	fTestCertPath string
	fTestKeyPath  string
	fAccessKey    string
	fSecretKey    string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
}

/*
Shell command to run this test:

	go test -v ./qiniu_sslcert_test.go -args \
	--QINIUSSLCERT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINIUSSLCERT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINIUSSLCERT_ACCESSKEY="your-access-key" \
	--QINIUSSLCERT_SECRETKEY="your-secret-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
		})
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
