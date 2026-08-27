//go:build tester

package ksyunslb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ksyun-slb"
)

var (
	fp               = tester.InitArgs("KSYUNSLB_")
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

	go test -tags=tester -v ./ksyun_slb_test.go -args \
	--KSYUNSLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--KSYUNSLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--KSYUNSLB_ACCESSKEYID="your-access-key-id" \
	--KSYUNSLB_SECRETACCESSKEY="your-secret-access-key" \
	--KSYUNSLB_REGION="cn-beijing-6"
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
