//go:build tester

package ctcccloudelb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ctcccloud-elb"
)

var (
	fp               = tester.InitArgs("CTCCCLOUDELB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fRegionId        string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fRegionId, "REGIONID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./ctcccloud_elb_test.go -args \
	--CTCCCLOUDELB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--CTCCCLOUDELB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--CTCCCLOUDELB_ACCESSKEYID="your-access-key-id" \
	--CTCCCLOUDELB_SECRETACCESSKEY="your-secret-access-key" \
	--CTCCCLOUDELB_REGIONID="your-region-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			RegionId:        fRegionId,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
