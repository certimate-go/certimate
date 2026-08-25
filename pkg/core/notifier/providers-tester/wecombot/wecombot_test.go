//go:build tester

package wecombot_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/wecombot"
)

var (
	fp          = tester.InitArgs("WECOMBOT_")
	fWebhookUrl string
)

func init() {
	fp.DefineString(&fWebhookUrl, "WEBHOOKURL")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./wecombot_test.go -args \
	--WECOMBOT_WEBHOOKURL="https://example.com/your-webhook-url" \
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			WebhookUrl: fWebhookUrl,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})
}
