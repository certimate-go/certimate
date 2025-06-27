package aliyuncas_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/ssl-manager/providers/aliyun-cas"
)

var (
	fInputCertPath   string
	fInputKeyPath    string
	fAccessKeyId     string
	fAccessKeySecret string
	fRegion          string
)

func init() {
	argsPrefix := "CERTIMATE_SSLMANAGER_ALIYUNCAS_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKeyId, argsPrefix+"ACCESSKEYID", "", "")
	flag.StringVar(&fAccessKeySecret, argsPrefix+"ACCESSKEYSECRET", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
}

/*
Shell command to run this test:

	go test -v ./aliyun_cas_test.go -args \
	--CERTIMATE_SSLMANAGER_ALIYUNCAS_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--CERTIMATE_SSLMANAGER_ALIYUNCAS_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--CERTIMATE_SSLMANAGER_ALIYUNCAS_ACCESSKEYID="your-access-key-id" \
	--CERTIMATE_SSLMANAGER_ALIYUNCAS_ACCESSKEYSECRET="your-access-key-secret" \
	--CERTIMATE_SSLMANAGER_ALIYUNCAS_REGION="cn-hangzhou"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYSECRET: %v", fAccessKeySecret),
			fmt.Sprintf("REGION: %v", fRegion),
		}, "\n"))

		sslmanager, err := provider.NewSSLManagerProvider(&provider.SSLManagerProviderConfig{
			AccessKeyId:     fAccessKeyId,
			AccessKeySecret: fAccessKeySecret,
			Region:          fRegion,
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
