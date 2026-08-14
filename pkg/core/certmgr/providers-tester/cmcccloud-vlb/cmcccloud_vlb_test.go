//go:build tester

package cmcccloudvlb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/cmcccloud-vlb"
)

var (
	fp               = tester.InitArgs("CMCCCLOUDVLB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fPoolId          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fPoolId, "POOLID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./cmcccloud_vlb_test.go -args \
	--CMCCCLOUDVLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CMCCCLOUDVLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CMCCCLOUDVLB_ACCESSKEYID="your-access-key-id" \
	--CMCCCLOUDVLB_ACCESSKEYSECRET="your-access-key-secret" \
	--CMCCCLOUDVLB_POOLID="CIDC-RP-29"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			PoolId:          fPoolId,
			IsDefault:       true,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
