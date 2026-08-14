//go:build tester

package discordbot_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/notifier/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/discordbot"
)

var (
	fp         = tester.InitArgs("DISCORDBOT_")
	fApiToken  string
	fChannelId string
)

func init() {
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fChannelId, "CHANNELID")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./discordbot_test.go -args \
	--DISCORDBOT_APITOKEN="your-bot-token" \
	--DISCORDBOT_CHANNELID="your-channel-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			BotToken:  fApiToken,
			ChannelId: fChannelId,
		})
		require.NoError(t, err)

		tester.Notify(t, provider, tester.NotifyInput{})
	})
}
