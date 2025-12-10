package cpanelsite_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/cpanel-site"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fUsername      string
	fApiToken      string
	fDomain        string
)

func init() {
	argsPrefix := "CPANELSITE_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fApiToken, argsPrefix+"APITOKEN", "", "")
	flag.StringVar(&fDomain, argsPrefix+"DOMAIN", "", "")
}

/*
Shell command to run this test:

	go test -v ./cpanel_site_test.go -args \
	--CPANELSITE_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--CPANELSITE_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--CPANELSITE_SERVERURL="http://127.0.0.1:2082" \
	--CPANELSITE_USERNAME="your-username" \
	--CPANELSITE_APITOKEN="your-api-token" \
	--CPANELSITE_DOMAIN="example.com"
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
			fmt.Sprintf("APITOKEN: %v", fApiToken),
			fmt.Sprintf("DOMAIN: %v", fDomain),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                fServerUrl,
			Username:                 fUsername,
			ApiToken:                 fApiToken,
			AllowInsecureConnections: true,
			Domain:                   fDomain,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		res, err := provider.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		t.Logf("ok: %v", res)
	})
}
