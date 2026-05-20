package aliyuncas_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-cas"
)

var (
	fp               = tester.Args("ALIYUNCAS_")
	fTestCertPath    string
	fTestKeyPath     string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fAccessKeyId, "ACCESSKEYID")
	fp.DefineString(&fAccessKeySecret, "ACCESSKEYSECRET")
	fp.DefineString(&fRegion, "REGION")
}

/*
Shell command to run this test:

	go test -v ./aliyun_cas_test.go -args \
	--ALIYUNCAS_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNCAS_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNCAS_ACCESSKEYID="your-access-key-id" \
	--ALIYUNCAS_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNCAS_REGION="cn-hangzhou"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
