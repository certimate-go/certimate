//go:build tester

package k8ssecret_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/deployer/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/deployer/providers/k8s-secret"
)

var (
	fp                   = tester.InitArgs("K8SSECRET_")
	fTestCertPath        string
	fTestKeyPath         string
	fKubeConfig          string
	fNamespace           string
	fSecretName          string
	fSecretDataKeyForCrt string
	fSecretDataKeyForKey string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fKubeConfig, "KUBECONFIG")
	fp.DefineString(&fNamespace, "NAMESPACE", "default")
	fp.DefineString(&fSecretName, "SECRETNAME")
	fp.DefineString(&fSecretDataKeyForCrt, "SECRETDATAKEYFORCRT", "tls.crt")
	fp.DefineString(&fSecretDataKeyForKey, "SECRETDATAKEYFORKEY", "tls.key")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./k8s_secret_test.go -args \
	--K8SSECRET_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--K8SSECRET_TESTKEYPATH="/path/to/your-test-key.pem" \
	--K8SSECRET_KUBECONFIG="..." \
	--K8SSECRET_NAMESPACE="default" \
	--K8SSECRET_SECRETNAME="secret" \
	--K8SSECRET_SECRETDATAKEYFORCRT="tls.crt" \
	--K8SSECRET_SECRETDATAKEYFORKEY="tls.key"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	if fKubeConfigStat, err := os.Stat(fKubeConfig); err == nil && !fKubeConfigStat.IsDir() {
		fKubeConfigBytes, _ := os.ReadFile(fKubeConfig)
		fKubeConfig = string(fKubeConfigBytes)
	}

	t.Run("Deploy", func(t *testing.T) {
		provider, err := impl.NewDeployer(&impl.DeployerConfig{
			KubeConfig:          fKubeConfig,
			Namespace:           fNamespace,
			SecretName:          fSecretName,
			SecretDataKeyForCrt: fSecretDataKeyForCrt,
			SecretDataKeyForKey: fSecretDataKeyForKey,
		})
		require.NoError(t, err)

		tester.Deploy(t, provider, tester.DeployInput{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
