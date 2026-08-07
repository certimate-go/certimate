package testing

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type TestDeployArgs struct {
	CertPath string
	KeyPath  string
}

func TestDeploy(t *testing.T, testProvider deployer.Provider, testArgs TestDeployArgs) {
	if _, err := os.Stat(testArgs.CertPath); os.IsNotExist(err) {
		t.Errorf("err: test cert file not exist")
		return
	}

	if _, err := os.Stat(testArgs.KeyPath); os.IsNotExist(err) {
		t.Errorf("err: test privkey file not exist")
		return
	}

	ctx := context.Background()
	certData, _ := os.ReadFile(testArgs.CertPath)
	privkeyData, _ := os.ReadFile(testArgs.KeyPath)
	assert.NotNil(t, certData)
	assert.NotNil(t, privkeyData)

	logger := slog.Default()
	logger.Enabled(ctx, slog.LevelDebug)
	testProvider.SetLogger(logger)

	res, err := testProvider.Deploy(ctx, string(certData), string(privkeyData))
	require.NoError(t, err)
	require.NotNil(t, res)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
