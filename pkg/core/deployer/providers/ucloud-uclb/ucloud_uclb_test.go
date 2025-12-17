package uclouduclb_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/ucloud-uclb"
)

var (
	fInputCertPath  string
	fInputKeyPath   string
	fPrivateKey     string
	fPublicKey      string
	fRegion         string
	fLoadbalancerId string
	fVServerId      string
)

func init() {
	argsPrefix := "UCLOUDUCLB_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fPrivateKey, argsPrefix+"PRIVATEKEY", "", "")
	flag.StringVar(&fPublicKey, argsPrefix+"PUBLICKEY", "", "")
	flag.StringVar(&fRegion, argsPrefix+"REGION", "", "")
	flag.StringVar(&fLoadbalancerId, argsPrefix+"LOADBALANCERID", "", "")
	flag.StringVar(&fVServerId, argsPrefix+"VSERVERID", "", "")
}

/*
Shell command to run this test:

	go test -v ./ucloud_uclb_test.go -args \
	--UCLOUDUCLB_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--UCLOUDUCLB_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--UCLOUDUCLB_PRIVATEKEY="your-private-key" \
	--UCLOUDUCLB_PUBLICKEY="your-public-key" \
	--UCLOUDUCLB_REGION="cn-bj2" \
	--UCLOUDUCLB_LOADBALANCERID="your-loadbalancer-id" \
	--UCLOUDUCLB_VSERVERID="your-vserver-id"
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
			fmt.Sprintf("VSERVERID: %v", fVServerId),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			PrivateKey:     fPrivateKey,
			PublicKey:      fPublicKey,
			Region:         fRegion,
			ResourceType:   provider.RESOURCE_TYPE_VSERVER,
			LoadbalancerId: fLoadbalancerId,
			VServerId:      fVServerId,
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
