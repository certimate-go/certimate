package larkbot_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/notifier/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/larkbot"
)

var (
	fp          = tester.Args("LARKBOT_")
	fWebhookUrl string
	fSecret     string
)

func init() {
	fp.DefineString(&fWebhookUrl, "WEBHOOKURL")
	fp.DefineString(&fSecret, "SECRET")
}

/*
Shell command to run this test:

	go test -v ./larkbot_test.go -args \
	--LARKBOT_WEBHOOKURL="https://example.com/your-webhook-url" \
	--LARKBOT_SECRET="your-secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			WebhookUrl: fWebhookUrl,
			Secret:     fSecret,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestNotify(t, provider, tester.TestNotifyArgs{})
	})
}
