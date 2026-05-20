package s3_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/s3"
)

var (
	fInputCertPath  string
	fInputKeyPath   string
	fSshHost        string
	fAccessKey      string
	fSecretKey      string
	fRegion         string
	fBucket         string
	fOutputCertPath string
	fOutputKeyPath  string
)

func init() {
	argsPrefix := "S3_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fSshHost, argsPrefix+"ENDPOINT", "", "")
	flag.StringVar(&fAccessKey, argsPrefix+"ACCESSKEY", "", "")
	flag.StringVar(&fSecretKey, argsPrefix+"SECRETKEY", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
	flag.StringVar(&fBucket, argsPrefix+"BUCKET", "", "")
	flag.StringVar(&fOutputCertPath, argsPrefix+"OUTPUTCERTPATH", "", "")
	flag.StringVar(&fOutputKeyPath, argsPrefix+"OUTPUTKEYPATH", "", "")
}

/*
Shell command to run this test:

	go test -v ./s3_test.go -args \
	--S3_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--S3_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--S3_ENDPOINT="http://endpoint" \
	--S3_ACCESSKEY="your-access-key" \
	--S3_SECRETKEY="your-secret-key" \
	--S3_REGION="your-region" \
	--S3_BUCKET="your-bucket" \
	--S3_OUTPUTCERTPATH="/path/to/your-output-cert.pem" \
	--S3_OUTPUTKEYPATH="/path/to/your-output-key.pem"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ENDPOINT: %v", fSshHost),
			fmt.Sprintf("ACCESSKEY: %v", fAccessKey),
			fmt.Sprintf("SECRETKEY: %v", fSecretKey),
			fmt.Sprintf("REGION: %v", fRegion),
			fmt.Sprintf("BUCKET: %v", fBucket),
			fmt.Sprintf("OUTPUTCERTPATH: %v", fOutputCertPath),
			fmt.Sprintf("OUTPUTKEYPATH: %v", fOutputKeyPath),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			Endpoint:            fSshHost,
			AccessKey:           fAccessKey,
			SecretKey:           fSecretKey,
			Region:              fRegion,
			Bucket:              fBucket,
			OutputFormat:        provider.OUTPUT_FORMAT_PEM,
			OutputCertObjectKey: fOutputCertPath,
			OutputKeyObjectKey:  fOutputKeyPath,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := provider.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
		}

		t.Logf("ok: %v", res)
	})
}
