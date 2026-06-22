package qingcloudlb_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/qingcloud-lb"
)

var (
	fp               = tester.Args("QINGCLOUDLB_")
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
