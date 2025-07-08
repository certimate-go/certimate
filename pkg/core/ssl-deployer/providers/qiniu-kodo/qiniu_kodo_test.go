package qiniukodo_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/ssl-deployer/providers/qiniu-kodo"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fAccessKey     string
	fSecretKey     string
	fDomain        string
)

func init() {
	argsPrefix := "CERTIMATE_SSLDEPLOYER_QINIUKODO_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKey, argsPrefix+"ACCESSKEY", "", "")
	flag.StringVar(&fSecretKey, argsPrefix+"SECRETKEY", "", "")
	flag.StringVar(&fDomain, argsPrefix+"DOMAIN", "", "")
}

/*
Shell command to run this test:

	go test -v ./qiniu_kodo_test.go -args \
	--CERTIMATE_SSLDEPLOYER_QINIUKODO_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--CERTIMATE_SSLDEPLOYER_QINIUKODO_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--CERTIMATE_SSLDEPLOYER_QINIUKODO_ACCESSKEY="your-access-key" \
	--CERTIMATE_SSLDEPLOYER_QINIUKODO_SECRETKEY="your-secret-key" \
	--CERTIMATE_SSLDEPLOYER_QINIUKODO_DOMAIN="example.com"
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
			fmt.Sprintf("DOMAIN: %v", fDomain),
		}, "\n"))

		deployer, err := provider.NewSSLDeployerProvider(&provider.SSLDeployerProviderConfig{
			AccessKey: fAccessKey,
			SecretKey: fSecretKey,
			Domain:    fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := deployer.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		t.Logf("ok: %v", res)
	})
}
