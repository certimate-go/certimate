package zenlayercdn_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-cdn"
)

var (
	fInputCertPath     string
	fInputKeyPath      string
	fAccessKeyId       string
	fAccessKeyPassword string
)

func init() {
	argsPrefix := "ZENLAYERCDN_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fAccessKeyId, argsPrefix+"ACCESSKEYID", "", "")
	flag.StringVar(&fAccessKeyPassword, argsPrefix+"ACCESSKEYPASSWORD", "", "")
}

/*
Shell command to run this test:

	go test -v ./zenlayer_cdn_test.go -args \
	--ZENLAYERCDN_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--ZENLAYERCDN_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--ZENLAYERCDN_ACCESSKEYID="your-access-key-id" \
	--ZENLAYERCDN_ACCESSKEYPASSWORD="your-secret-access-key"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("ACCESSKEYID: %v", fAccessKeyId),
			fmt.Sprintf("ACCESSKEYPASSWORD: %v", fAccessKeyPassword),
		}, "\n"))

		provider, err := provider.NewCertmgr(&provider.CertmgrConfig{
			AccessKeyId:       fAccessKeyId,
			AccessKeyPassword: fAccessKeyPassword,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := provider.Upload(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		sres, _ := json.Marshal(res)
		t.Logf("ok: %s", string(sres))
	})
}
