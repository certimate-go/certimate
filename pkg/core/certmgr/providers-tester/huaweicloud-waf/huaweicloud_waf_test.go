//go:build tester

package huaweicloudwaf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/huaweicloud-waf"
)

var (
	fp               = tester.InitArgs("HUAWEICLOUDWAF_")
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

	go test -tags=tester -v ./huaweicloud_waf_test.go -args \
	--HUAWEICLOUDWAF_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--HUAWEICLOUDWAF_TESTKEYPATH="/path/to/your-test-key.pem" \
	--HUAWEICLOUDWAF_ACCESSKEYID="your-access-key-id" \
	--HUAWEICLOUDWAF_SECRETACCESSKEY="your-access-key-secret" \
	--HUAWEICLOUDWAF_REGION="cn-north-4"
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
