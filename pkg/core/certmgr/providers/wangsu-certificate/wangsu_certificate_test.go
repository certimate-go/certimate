package wangsucertificate_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/wangsu-certificate"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp               = it.Args("WANGSUCERTIFICATE_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
}

/*
Shell command to run this test:

	go test -v ./wangsu_certificate_test.go -args \
	--WANGSUCERTIFICATE_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--WANGSUCERTIFICATE_TESTKEYPATH="/path/to/your-test-key.pem" \
	--WANGSUCERTIFICATE_ACCESSKEYID="your-access-key-id" \
	--WANGSUCERTIFICATE_ACCESSKEYSECRET="your-access-key-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
		})
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
