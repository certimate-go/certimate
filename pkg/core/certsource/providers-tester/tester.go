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

	"github.com/certimate-go/certimate/pkg/core/certsource"
)

type FetchInput struct {
	Domain string
	CertId string
}

func Fetch(t *testing.T, provider certsource.Provider, input FetchInput) {
	if input.Domain == "" && input.CertId == "" {
		t.Errorf("err: test domain and certId are both empty")
		return
	}

	ctx := context.Background()

	loglvr := slog.LevelVar{}
	loglvr.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &loglvr}))
	provider.SetLogger(logger)

	res, err := provider.Fetch(ctx, input.Domain, input.CertId)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.NotEmpty(t, res.CertPEM)
	assert.NotEmpty(t, res.PrivkeyPEM)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
