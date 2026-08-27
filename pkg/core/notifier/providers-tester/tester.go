//go:build tester

package tester

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/certimate-go/certimate/pkg/core/notifier"
)

const (
	mockSubject = "test_subject"
	mockMessage = "test_message"
)

type NotifyInput struct {
	Subject string
	Message string
}

func Notify(t *testing.T, provider notifier.Provider, input NotifyInput) {
	ctx := context.Background()
	message := lo.Ternary(input.Message != "", input.Message, mockMessage)
	subject := lo.Ternary(input.Subject != "", input.Subject, mockSubject)

	logger := slog.Default()
	logger.Enabled(ctx, slog.LevelDebug)
	provider.SetLogger(logger)

	res, err := provider.Notify(ctx, message, subject)
	require.NoError(t, err)
	require.NotNil(t, res)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
