//go:build !windows
// +build !windows

package cmd

import (
	"ganigeorgiev.com/pocketbase/pocketbase"
)

func Serve(app *pocketbase.PocketBase) error {
	return app.Start()
}
