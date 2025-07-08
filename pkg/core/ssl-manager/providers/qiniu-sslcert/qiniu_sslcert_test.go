package qiniusslcert_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/qiniu-sslcert"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fAccessKey     string
	fSecretKey     string
)

func init() {
	argsPrefix := "CERTIMATE_SSLMANAGER_QINIUSSLCERT_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKey, argsPrefix+"ACCESSKEY", "", "")
	flag.StringVar(&fSecretKey, argsPrefix+"SECRETKEY", "", "")
}

/*
Shell command to run this test:

	go test -v ./qiniu_sslcert_test.go -args \
	--CERTIMATE_SSLMANAGER_QINIUSSLCERT_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--CERTIMATE_SSLMANAGER_QINIUSSLCERT_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--CERTIMATE_SSLMANAGER_QINIUSSLCERT_ACCESSKEY="your-access-key" \
	--CERTIMATE_SSLMANAGER_QINIUSSLCERT_SECRETKEY="your-secret-key"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEY: %v", fAccessKey),
			fmt.Sprintf("SECRETKEY: %v", fSecretKey),
		}, "\n"))

		sslmanager, err := provider.NewSSLManagerProvider(&provider.SSLManagerProviderConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := sslmanager.Upload(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		sres, _ := json.Marshal(res)
		t.Logf("ok: %s", string(sres))
	})
}
