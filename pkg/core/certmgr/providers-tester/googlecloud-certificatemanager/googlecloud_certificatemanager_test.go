//go:build tester

package googlecloudcertificatemanager_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/certmgr/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/googlecloud-certificatemanager"
)

var (
	fp                 = tester.InitArgs("GOOGLECLOUDCERTIFICATEMANAGER_")
	fTestCertPath      string
	fTestKeyPath       string
	fProjectId         string
	fServiceAccountKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fProjectId, "PROJECTID")
	fp.DefineString(&fServiceAccountKey, "SERVICEACCOUNTKEY")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./googlecloud_certificatemanager_test.go -args \
	--GOOGLECLOUDCERTIFICATEMANAGER_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--GOOGLECLOUDCERTIFICATEMANAGER_TESTKEYPATH="/path/to/your-test-key.pem" \
	--GOOGLECLOUDCERTIFICATEMANAGER_PROJECTID="your-project-id" \
	--GOOGLECLOUDCERTIFICATEMANAGER_SERVICEACCOUNTKEY="{...}"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	if fServiceAccountKeyStat, err := os.Stat(fServiceAccountKey); err == nil && !fServiceAccountKeyStat.IsDir() {
		fServiceAccountKeyBytes, _ := os.ReadFile(fServiceAccountKey)
		fServiceAccountKey = string(fServiceAccountKeyBytes)
	}

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			ProjectId:         fProjectId,
			ServiceAccountKey: fServiceAccountKey,
		})
		require.NoError(t, err)

		tester.Upload(t, provider, tester.UploadInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
