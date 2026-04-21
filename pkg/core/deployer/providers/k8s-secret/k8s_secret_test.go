package k8ssecret_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/k8s-secret"
)

var (
	fInputCertPath          string
	fInputKeyPath           string
	fInputIssuerPath        string
	fNamespace              string
	fSecretName             string
	fSecretDataKeyForCrt    string
	fSecretDataKeyForKey    string
	fSecretDataKeyForIssuer string
)

func init() {
	argsPrefix := "K8SSECRET_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "")
	flag.StringVar(&fInputIssuerPath, argsPrefix+"INPUTISSUERPATH", "", "")
	flag.StringVar(&fNamespace, argsPrefix+"NAMESPACE", "default", "")
	flag.StringVar(&fSecretName, argsPrefix+"SECRETNAME", "", "")
	flag.StringVar(&fSecretDataKeyForCrt, argsPrefix+"SECRETDATAKEYFORCRT", "tls.crt", "")
	flag.StringVar(&fSecretDataKeyForKey, argsPrefix+"SECRETDATAKEYFORKEY", "tls.key", "")
	flag.StringVar(&fSecretDataKeyForIssuer, argsPrefix+"SECRETDATAKEYFORISSUER", "issuer.crt", "")
}

/*
Shell command to run this test:

	go test -v ./k8s_secret_test.go -args \
	--K8SSECRET_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--K8SSECRET_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--K8SSECRET_INPUTISSUERPATH="/path/to/your-input-issuer.pem" \
	--K8SSECRET_NAMESPACE="default" \
	--K8SSECRET_SECRETNAME="secret" \
	--K8SSECRET_SECRETDATAKEYFORCRT="tls.crt" \
	--K8SSECRET_SECRETDATAKEYFORKEY="tls.key" \
	--K8SSECRET_SECRETDATAKEYFORISSUER="issuer.crt"
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("NAMESPACE: %v", fNamespace),
			fmt.Sprintf("SECRETNAME: %v", fSecretName),
			fmt.Sprintf("SECRETDATAKEYFORCRT: %v", fSecretDataKeyForCrt),
			fmt.Sprintf("SECRETDATAKEYFORKEY: %v", fSecretDataKeyForKey),
			fmt.Sprintf("SECRETDATAKEYFORISSUER: %v", fSecretDataKeyForIssuer),
		}, "\n"))

		provider, err := provider.NewDeployer(&provider.DeployerConfig{
			Namespace:              fNamespace,
			SecretName:             fSecretName,
			SecretDataKeyForCrt:    fSecretDataKeyForCrt,
			SecretDataKeyForKey:    fSecretDataKeyForKey,
			SecretDataKeyForIssuer: fSecretDataKeyForIssuer,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		fInputCertData, _ := os.ReadFile(fInputCertPath)
		fInputKeyData, _ := os.ReadFile(fInputKeyPath)
		fInputIssuerData, _ := os.ReadFile(fInputIssuerPath)
		res, err := provider.Deploy(context.Background(), string(fInputCertData), string(fInputKeyData), string(fInputIssuerData))
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		t.Logf("ok: %v", res)
	})
}
