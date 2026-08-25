//go:build tester

package tester

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

type DeployInput struct {
	CertPath string
	KeyPath  string
}

func Deploy(t *testing.T, provider deployer.Provider, input DeployInput) {
	if _, err := os.Stat(input.CertPath); os.IsNotExist(err) {
		t.Errorf("err: test cert file not exist")
		return
	}

	if _, err := os.Stat(input.KeyPath); os.IsNotExist(err) {
		t.Errorf("err: test privkey file not exist")
		return
	}

	ctx := context.Background()
	certData, _ := os.ReadFile(input.CertPath)
	privkeyData, _ := os.ReadFile(input.KeyPath)
	assert.NotNil(t, certData)
	assert.NotNil(t, privkeyData)

	logger := slog.Default()
	logger.Enabled(ctx, slog.LevelDebug)
	provider.SetLogger(logger)

	res, err := provider.Deploy(ctx, string(certData), string(privkeyData))
	require.NoError(t, err)
	require.NotNil(t, res)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
