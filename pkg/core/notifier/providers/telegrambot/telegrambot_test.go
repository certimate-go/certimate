package telegrambot_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/notifier/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/telegrambot"
)

var (
	fp        = tester.Args("TELEGRAMBOT_")
	fApiToken string
	fChatId   string
)

func init() {
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fChatId, "CHATID")
}

/*
Shell command to run this test:

	go test -v ./telegrambot_test.go -args \
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
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestNotify(t, provider, tester.TestNotifyArgs{})
	})
}
