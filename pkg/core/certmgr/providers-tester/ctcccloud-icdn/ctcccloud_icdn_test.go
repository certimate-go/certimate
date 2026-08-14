//go:build tester

package ctcccloudicdn_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-icdn"
)

var (
	fp               = tester.InitArgs("CTCCCLOUDICDN_")
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

	go test -tags=tester -v ./ctcccloud_icdn_test.go -args \
	--CTCCCLOUDICDN_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDICDN_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDICDN_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDICDN_SECRETACCESSKEY="your-secret-access-key"
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
