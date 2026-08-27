//go:build tester

package baiducloudcert_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/baiducloud-cert"
)

var (
	fp               = tester.InitArgs("BAIDUCLOUDCERT_")
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

	go test -tags=tester -v ./baiducloud_cert_test.go -args \
	--BAIDUCLOUDCERT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--BAIDUCLOUDCERT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--BAIDUCLOUDCERT_ACCESSKEYID="your-access-key-id" \
	--BAIDUCLOUDCERT_SECRETACCESSKEY="your-access-key-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
