package aliyunslb_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aliyun-slb"
)

var (
	fp               = tester.Args("ALIYUNSLB_")
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

	go test -v ./aliyun_slb_test.go -args \
	--ALIYUNSLB_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ALIYUNSLB_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ALIYUNSLB_ACCESSKEYID="your-access-key-id" \
	--ALIYUNSLB_ACCESSKEYSECRET="your-access-key-secret" \
	--ALIYUNSLB_REGION="cn-hangzhou"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("TESTCERTPATH: %v", fTestCertPath),
			fmt.Sprintf("TESTKEYPATH: %v", fTestKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYSECRET: %v", fAccessKeySecret),
			fmt.Sprintf("REGION: %v", fRegion),
		}, "\n"))

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
