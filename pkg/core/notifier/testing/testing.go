package testing

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

type TestNotifyArgs struct {
	Subject string
	Message string
}

func TestNotify(t *testing.T, testProvider notifier.Provider, testArgs TestNotifyArgs) {
	ctx := context.Background()
	message := lo.Ternary(testArgs.Message != "", testArgs.Message, mockMessage)
	subject := lo.Ternary(testArgs.Subject != "", testArgs.Subject, mockSubject)

	logger := slog.Default()
	logger.Enabled(ctx, slog.LevelDebug)
	testProvider.SetLogger(logger)

	res, err := testProvider.Notify(ctx, message, subject)
	require.NoError(t, err)
	require.NotNil(t, res)

	resjson, _ := json.Marshal(res)
	t.Logf("ok: %s", string(resjson))
}
