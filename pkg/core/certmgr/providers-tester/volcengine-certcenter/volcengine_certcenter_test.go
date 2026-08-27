//go:build tester

package volcenginecertcenter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/volcengine-certcenter"
)

var (
	fp               = tester.InitArgs("VOLCENGINECERTCENTER_")
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

	go test -tags=tester -v ./volcengine_certcenter_test.go -args \
	--VOLCENGINECERTCENTER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--VOLCENGINECERTCENTER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--VOLCENGINECERTCENTER_ACCESSKEYID="your-access-key-id" \
	--VOLCENGINECERTCENTER_SECRETACCESSKEY="your-secret-access-key"
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
