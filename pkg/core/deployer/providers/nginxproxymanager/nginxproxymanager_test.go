package nginxproxymanager_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/nginxproxymanager"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fServerUrl     string
	fUsername      string
	fPassword      string
	fHostType      string
	fHostId        int64
)

func init() {
	argsPrefix := "NGINXPROXYMANAGER_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fServerUrl, argsPrefix+"SERVERURL", "", "")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "")
	flag.StringVar(&fHostType, argsPrefix+"HOSTTYPE", "", "")
	flag.Int64Var(&fHostId, argsPrefix+"HOSTID", 0, "")
}

/*
Shell command to run this test:

	go test -v ./nginxproxymanager_test.go -args \
	--NGINXPROXYMANAGER_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--NGINXPROXYMANAGER_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--NGINXPROXYMANAGER_SERVERURL="http://127.0.0.1:20410" \
	--NGINXPROXYMANAGER_USERNAME="your-username" \
	--NGINXPROXYMANAGER_PASSWORD="your-password" \
	--NGINXPROXYMANAGER_HOSTTYPE="proxy" \
	--NGINXPROXYMANAGER_HOSTID="your-host-id"
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
			fmt.Sprintf("HOSTTYPE: %v", fHostType),
			fmt.Sprintf("HOSTID: %v", fHostId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			ServerUrl:                fServerUrl,
			AuthMethod:               provider.AUTH_METHOD_PASSWORD,
			Username:                 fUsername,
			Password:                 fPassword,
			AllowInsecureConnections: true,
			ResourceType:             provider.RESOURCE_TYPE_HOST,
			HostType:                 fHostType,
			HostMatchPattern:         provider.HOST_MATCH_PATTERN_SPECIFIED,
			HostId:                   fHostId,
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
