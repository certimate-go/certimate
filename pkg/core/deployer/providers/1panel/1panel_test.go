package onepanel_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/1panel"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fApiVersion    string
	fApiKey        string
	fWebsiteId     int64
)

func init() {
	argsPrefix := "1PANEL_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fApiVersion, argsPrefix+"APIVERSION", "v1", "")
	flag.StringVar(&fApiKey, argsPrefix+"APIKEY", "", "")
	flag.Int64Var(&fWebsiteId, argsPrefix+"WEBSITEID", 0, "")
}

/*
Shell command to run this test:

	go test -v ./1panel_test.go -args \
	--1PANEL_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--1PANEL_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--1PANEL_SERVERURL="http://127.0.0.1:20410" \
	--1PANEL_APIVERSION="v1" \
	--1PANEL_APIKEY="your-api-key" \
	--1PANEL_WEBSITEID="your-website-id"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("SERVERURL: %v", fServerUrl),
			fmt.Sprintf("APIVERSION: %v", fApiVersion),
			fmt.Sprintf("APIKEY: %v", fApiKey),
			fmt.Sprintf("WEBSITEID: %v", fWebsiteId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                fServerUrl,
			ApiVersion:               fApiVersion,
			ApiKey:                   fApiKey,
			AllowInsecureConnections: true,
			ResourceType:             provider.RESOURCE_TYPE_WEBSITE,
			WebsiteMatchPattern:      provider.WEBSITE_MATCH_PATTERN_SPECIFIED,
			WebsiteId:                fWebsiteId,
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
