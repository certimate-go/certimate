package matrix_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/notifier/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/notifier/providers/matrix"
)

var (
	fp             = tester.Args("MATRIX_")
	fHomeserverURL string
	fRoomID        string
	fAccessToken   string
	fUserID        string
	fPassword      string
	fAuthMode      string
)

func init() {
	fp.DefineString(&fHomeserverURL, "HOMESERVERURL")
	fp.DefineString(&fRoomID, "ROOMID")
	fp.DefineString(&fAccessToken, "ACCESSTOKEN")
	fp.DefineString(&fUserID, "USERID")
	fp.DefineString(&fPassword, "PASSWORD")
	fp.DefineString(&fAuthMode, "AUTHMODE")
}

/*
Shell command to run integration test:

	go test -v ./matrix_test.go -args \
	--MATRIX_HOMESERVERURL="https://matrix.example.org" \
	--MATRIX_ROOMID="!room:matrix.example.org" \
	--MATRIX_ACCESSTOKEN="syt_..." \
	--MATRIX_AUTHMODE="token"
*/
func TestProvider(t *testing.T) {
	fp.Parse()
	if fHomeserverURL == "" || fRoomID == "" {
		t.Skip("set MATRIX_HOMESERVERURL and MATRIX_ROOMID for integration test")
	}

	t.Run("Notify", func(t *testing.T) {
		cfg := &impl.NotifierConfig{
			HomeserverUrl: fHomeserverURL,
			RoomId:        fRoomID,
			AuthMode:      fAuthMode,
			AccessToken:   fAccessToken,
			UserId:        fUserID,
			Password:      fPassword,
		}
		if cfg.AuthMode == "" {
			if cfg.AccessToken != "" {
				cfg.AuthMode = "token"
			} else {
				cfg.AuthMode = "password"
			}
		}

		provider, err := impl.NewNotifier(cfg)
		if err != nil {
			t.Fatalf("NewNotifier: %v", err)
		}

		tester.TestNotify(t, provider, tester.TestNotifyArgs{})
	})
}
