package slackbot_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/slackbot"
	it "github.com/certimate-go/certimate/pkg/core/notifier/testing"
)

var (
	fp         = it.Args("SLACKBOT_")
	fApiToken  string
	fChannelId string
)

func init() {
	fp.DefineString(&fApiToken, "APITOKEN")
	fp.DefineString(&fChannelId, "CHANNELID")
}

/*
Shell command to run this test:

	go test -v ./slackbot_test.go -args \
	--SLACKBOT_APITOKEN="your-bot-token" \
	--SLACKBOT_CHANNELID="your-channel-id"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			BotToken:  fApiToken,
			ChannelId: fChannelId,
		})
		require.NoError(t, err)

		it.TestNotify(t, provider, it.TestNotifyArgs{})
	})
}
