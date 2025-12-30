package ratpanel_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/ratpanel"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fAccessTokenId int64
	fAccessToken   string
	fSiteName      string
)

func init() {
	argsPrefix := "RATPANEL_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.Int64Var(&fAccessTokenId, argsPrefix+"ACCESSTOKENID", 0, "")
	flag.StringVar(&fAccessToken, argsPrefix+"ACCESSTOKEN", "", "")
	flag.StringVar(&fSiteName, argsPrefix+"SITENAME", "", "")
}

/*
Shell command to run this test:

	go test -v ./ratpanel_test.go -args \
	--RATPANEL_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--RATPANEL_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--RATPANEL_SERVERURL="http://127.0.0.1:8888" \
	--RATPANEL_ACCESSTOKENID="your-access-token-id" \
	--RATPANEL_ACCESSTOKEN="your-access-token" \
	--RATPANEL_SITENAME="your-site-name"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("SERVERURL: %v", fServerUrl),
			fmt.Sprintf("ACCESSTOKENID: %v", fAccessTokenId),
			fmt.Sprintf("ACCESSTOKEN: %v", fAccessToken),
			fmt.Sprintf("SITENAME: %v", fSiteName),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                fServerUrl,
			AccessTokenId:            fAccessTokenId,
			AccessToken:              fAccessToken,
			AllowInsecureConnections: true,
			ResourceType:             provider.RESOURCE_TYPE_WEBSITE,
			SiteNames:                []string{fSiteName},
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
