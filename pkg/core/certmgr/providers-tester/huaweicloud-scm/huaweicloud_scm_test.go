//go:build tester

package huaweicloudscm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-scm"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDSCM_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegion          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegion, "REGION")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./huaweicloud_scm_test.go -args \
	--HUAWEICLOUDSCM_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDSCM_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDSCM_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDSCM_SECRETACCESSKEY="your-access-key-secret" \
	--HUAWEICLOUDSCM_REGION="cn-north-4"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			Region:          fRegion,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
