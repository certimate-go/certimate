package s3_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/deployer/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/s3"
)

var (
	fp               = tester.Args("S3_")
	fTestCertPath    string
	fTestKeyPath     string
	fSshHost         string
	fAccessKey       string
	fSecretKey       string
	fRegion          string
	fBucket          string
	fObjectKeyForCrt string
	fObjectKeyForKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fSshHost, "ENDPOINT")
	fp.DefineString(&fAccessKey, "ACCESSKEY")
	fp.DefineString(&fSecretKey, "SECRETKEY")
	fp.DefineString(&fRegion, "REGION")
	fp.DefineString(&fBucket, "BUCKET")
	fp.DefineString(&fObjectKeyForCrt, "OBJECTKEYFORCRT")
	fp.DefineString(&fObjectKeyForKey, "OBJECTKEYFORKEY")
}

/*
Shell command to run this test:

	go test -v ./s3_test.go -args \
	--S3_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--S3_TESTKEYPATH="/path/to/your-test-key.pem" \
	--S3_ENDPOINT="http://endpoint" \
	--S3_ACCESSKEY="your-access-key" \
	--S3_SECRETKEY="your-secret-key" \
	--S3_REGION="your-region" \
	--S3_BUCKET="your-bucket" \
	--S3_OBJECTKEYFORCRT="/path/to/your-output-cert.pem" \
	--S3_OBJECTKEYFORKEY="/path/to/your-output-key.pem"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Deploy_PEM", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			Endpoint:        fSshHost,
			AccessKey:       fAccessKey,
			SecretKey:       fSecretKey,
			Region:          fRegion,
			Bucket:          fBucket,
			FileFormat:      impl.FILE_FORMAT_PEM,
			ObjectKeyForCrt: fObjectKeyForCrt + ".pem",
			ObjectKeyForKey: fObjectKeyForKey + ".pem",
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestDeploy(t, provider, tester.TestDeployArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
