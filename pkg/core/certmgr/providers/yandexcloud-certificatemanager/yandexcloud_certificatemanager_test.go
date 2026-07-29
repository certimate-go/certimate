package yandexcloudcertificatemanager_test

import (
	"os"
	"testing"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/yandexcloud-certificatemanager"
	tester "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp                 = tester.Args("YANDEXCLOUDCERTIFICATEMANAGER_")
	fTestCertPath      string
	fTestKeyPath       string
	fFolderId          string
	fServiceAccountKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fFolderId, "FOLDERID")
	fp.DefineString(&fServiceAccountKey, "SERVICEACCOUNTKEY")
}

/*
Shell command to run this test:

	go test -v ./yandexcloud_certificatemanager_test.go -args \
	--YANDEXCLOUDCERTIFICATEMANAGER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--YANDEXCLOUDCERTIFICATEMANAGER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--YANDEXCLOUDCERTIFICATEMANAGER_FOLDERID="your-folder-id" \
	--YANDEXCLOUDCERTIFICATEMANAGER_SERVICEACCOUNTKEY="{...}"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	if fServiceAccountKeyStat, err := os.Stat(fServiceAccountKey); err == nil && !fServiceAccountKeyStat.IsDir() {
		fServiceAccountKeyBytes, _ := os.ReadFile(fServiceAccountKey)
		fServiceAccountKey = string(fServiceAccountKeyBytes)
	}

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			FolderId:          fFolderId,
			ServiceAccountKey: fServiceAccountKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
