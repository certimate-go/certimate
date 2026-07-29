package ctcccloudcdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-cdn"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp               = it.Args("CTCCCLOUDCDN_")
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

	go test -v ./ctcccloud_cdn_test.go -args \
	--CTCCCLOUDCDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDCDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDCDN_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDCDN_SECRETACCESSKEY="your-secret-access-key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
		})
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
