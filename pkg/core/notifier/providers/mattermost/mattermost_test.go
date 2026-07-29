package mattermost_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/mattermost"
	it "github.com/certimate-go/certimate/pkg/core/notifier/testing"
)

var (
	fp         = it.Args("MATTERMOST_")
	fServerUrl string
	fChannelId string
	fUsername  string
	fPassword  string
)

func init() {
	fp.DefineString(&fServerUrl, "SERVERURL")
	fp.DefineString(&fChannelId, "CHANNELID")
	fp.DefineString(&fUsername, "USERNAME")
	fp.DefineString(&fPassword, "PASSWORD")
}

/*
Shell command to run this test:

	go test -v ./mattermost_test.go -args \
	--MATTERMOST_SERVERURL="https://example.com/your-server-url" \
	--MATTERMOST_CHANNELID="your-chanel-id" \
	--MATTERMOST_USERNAME="your-username" \
	--MATTERMOST_PASSWORD="your-password"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Notify", func(t *testing.T) {
		provider, err := impl.NewNotifier(&impl.NotifierConfig{
			ServerUrl: fServerUrl,
			ChannelId: fChannelId,
			Username:  fUsername,
			Password:  fPassword,
		})
		require.NoError(t, err)

		it.TestNotify(t, provider, it.TestNotifyArgs{})
	})
}
