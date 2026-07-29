package qingcloudlb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/qingcloud-lb"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp               = it.Args("QINGCLOUDLB_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fSecretAccessKey string
	fZoneId          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fSecretAccessKey, "SECRETACCESSKEY")
	fp.DefineString(&fZoneId, "ZONEID")
}

/*
Shell command to run this test:

	go test -v ./qingcloud_lb_test.go -args \
	--QINGCLOUDLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--QINGCLOUDLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--QINGCLOUDLB_ACCESSKEYID="your-access-key-id" \
	--QINGCLOUDLB_SECRETACCESSKEY="your-secret-access-key" \
	--QINGCLOUDLB_ZONEID="pek3a"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			SecretAccessKey: fSecretAccessKey,
			ZoneId:          fZoneId,
		})
		require.NoError(t, err)

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
