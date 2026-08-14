//go:build tester

package telegrambot_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/telegrambot"
)

var (
	fp        = tester.InitArgs("TELEGRAMBOT_")
	fApiToken string
	fChatId   string
)

func init() {
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fChatId, "CHATID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./telegrambot_test.go -args \
	--TELEGRAMBOT_APITOKEN="your-api-token" \
	--TELEGRAMBOT_CHATID="your-chat-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			BotToken: fApiToken,
			ChatId:   fChatId,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})
}
