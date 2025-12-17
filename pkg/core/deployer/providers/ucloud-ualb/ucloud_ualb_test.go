package ucloudualb_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-ualb"
)

var (
	fInputCertPath  string
	fInputKeyPath   string
	fPrivateKey     string
	fPublicKey      string
	fRegion         string
	fLoadbalancerId string
	fListenerId     string
)

func init() {
	argsPrefix := "UCLOUDUALB_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fPrivateKey, argsPrefix+"PRIVATEKEY", "", "")
	flag.StringVar(&fPublicKey, argsPrefix+"PUBLICKEY", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
	flag.StringVar(&fLoadbalancerId, argsPrefix+"LOADBALANCERID", "", "")
	flag.StringVar(&fListenerId, argsPrefix+"LISTENERID", "", "")
}

/*
Shell command to run this test:

	go test -v ./ucloud_ualb_test.go -args \
	--UCLOUDUALB_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--UCLOUDUALB_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--UCLOUDUALB_PRIVATEKEY="your-private-key" \
	--UCLOUDUALB_PUBLICKEY="your-public-key" \
	--UCLOUDUALB_REGION="cn-bj2" \
	--UCLOUDUALB_LOADBALANCERID="your-loadbalancer-id" \
	--UCLOUDUALB_LISTENERID="your-listener-id"
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
			fmt.Sprintf("LOADBALANCERID: %v", fLoadbalancerId),
			fmt.Sprintf("LISTENERID: %v", fListenerId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			PrivateKey:     fPrivateKey,
			PublicKey:      fPublicKey,
			Region:         fRegion,
			ResourceType:   provider.RESOURCE_TYPE_LISTENER,
			LoadbalancerId: fLoadbalancerId,
			ListenerId:     fListenerId,
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
