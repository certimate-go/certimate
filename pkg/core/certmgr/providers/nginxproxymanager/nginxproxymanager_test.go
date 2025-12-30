package nginxproxymanager_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/certmgr/providers/nginxproxymanager"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fUsername      string
	fPassword      string
)

func init() {
	argsPrefix := "NGINXPROXYMANAGER_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fPassword, argsPrefix+"PASSWORD", "", "")
}

/*
Shell command to run this test:

	go test -v ./nginxproxymanager_test.go -args \
	--NGINXPROXYMANAGER_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--NGINXPROXYMANAGER_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--NGINXPROXYMANAGER_SERVERURL="http://127.0.0.1:81" \
	--NGINXPROXYMANAGER_USERNAME="your-username" \
	--NGINXPROXYMANAGER_PASSWORD="your-password"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("SERVERURL: %v", fServerUrl),
			fmt.Sprintf("USERNAME: %v", fUsername),
			fmt.Sprintf("PASSWORD: %v", fPassword),
		}, "\n"))

		provider, err := provider.NewCertmgr(&provider.CertmgrConfig{
			ServerUrl:                fServerUrl,
			AuthMethod:               provider.AUTH_METHOD_PASSWORD,
			Username:                 fUsername,
			Password:                 fPassword,
			AllowInsecureConnections: true,
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
