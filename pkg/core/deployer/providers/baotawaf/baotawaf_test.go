package baotawaf_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/baotawaf"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fApiKey        string
	fSiteName      string
	fSitePort      int64
)

func init() {
	argsPrefix := "BAOTAWAF_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fApiKey, argsPrefix+"APIKEY", "", "")
	flag.StringVar(&fSiteName, argsPrefix+"SITENAME", "", "")
	flag.Int64Var(&fSitePort, argsPrefix+"SITEPORT", 0, "")
}

/*
Shell command to run this test:

	go test -v ./baotawaf_test.go -args \
	--BAOTAWAF_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--BAOTAWAF_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--BAOTAWAF_SERVERURL="http://127.0.0.1:8888" \
	--BAOTAWAF_APIKEY="your-api-key" \
	--BAOTAWAF_SITENAME="your-site-name" \
	--BAOTAWAF_SITEPORT=443
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("SERVERURL: %v", fServerUrl),
			fmt.Sprintf("APIKEY: %v", fApiKey),
			fmt.Sprintf("SITENAME: %v", fSiteName),
			fmt.Sprintf("SITEPORT: %v", fSitePort),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			SiteNames:                []string{fSiteName},
			SitePort:                 int32(fSitePort),
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
