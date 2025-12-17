package ucloudulb_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ulb"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fPrivateKey    string
	fPublicKey     string
	fRegion        string
)

func init() {
	argsPrefix := "UCLOUDULB_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fPrivateKey, argsPrefix+"PRIVATEKEY", "", "")
	flag.StringVar(&fPublicKey, argsPrefix+"PUBLICKEY", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ulb_test.go -args \
	--UCLOUDULB_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--UCLOUDULB_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--UCLOUDULB_PRIVATEKEY="your-private-key" \
	--UCLOUDULB_PUBLICKEY="your-public-key" \
	--UCLOUDULB_REGION="cn-bj2"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("PRIVATEKEY: %v", fPrivateKey),
			fmt.Sprintf("PUBLICKEY: %v", fPublicKey),
			fmt.Sprintf("REGION: %v", fRegion),
		}, "\n"))

		provider, err := provider.NewCertmgr(&provider.CertmgrConfig{
			PrivateKey: fPrivateKey,
			PublicKey:  fPublicKey,
			Region:     fRegion,
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
