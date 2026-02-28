package flyio_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/flyio"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fApiToken      string
	fAppName       string
	fDomain        string
)

func init() {
	argsPrefix := "FLYIO_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fApiToken, argsPrefix+"APITOKEN", "", "")
	flag.StringVar(&fAppName, argsPrefix+"APPNAME", "", "")
	flag.StringVar(&fDomain, argsPrefix+"DOMAIN", "", "")
}

/*
Shell command to run this test:

	go test -v ./flyio_test.go -args \
	--FLYIO_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--FLYIO_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--FLYIO_APITOKEN="your-api-token" \
	--FLYIO_APPNAME="your-app-name" \
	--FLYIO_DOMAIN="example.com"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("APITOKEN: %v", fApiToken),
			fmt.Sprintf("APPNAME: %v", fAppName),
			fmt.Sprintf("DOMAIN: %v", fDomain),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ApiToken: fApiToken,
			AppName:  fAppName,
			Domain:   fDomain,
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
