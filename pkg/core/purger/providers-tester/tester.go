//go:build tester

package tester

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/certimate-go/certimate/pkg/core/purger"
)

type PurgeInput struct {
	Expiry time.Duration
}

func Purge(t *testing.T, provider purger.Provider, input PurgeInput) {
	ctx := context.Background()

	loglvr := slog.LevelVar{}
	loglvr.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &loglvr}))
	provider.SetLogger(logger)

	res, err := provider.Purge(ctx, input.Expiry)
	require.NoError(t, err)
	require.NotNil(t, res)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
