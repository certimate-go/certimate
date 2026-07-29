package awsiam_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-iam"
	it "github.com/certimate-go/certimate/pkg/core/certmgr/testing"
)

var (
	fp               = it.Args("AWSIAM_")
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

	go test -v ./aws_iam_test.go -args \
	--AWSIAM_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AWSIAM_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AWSIAM_ACCESSKEYID="your-access-key-id" \
	--AWSIAM_SECRETACCESSKEY="your-access-key-secret" \
	--AWSIAM_REGION="us-east-1"
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

		it.TestUpload(t, provider, it.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
