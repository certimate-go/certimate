package ui

import (
	"embed"

	"github.com/pocketbase/pocketbase/apis"
)

//go:embed all:dist
var distDir embed.FS

var DistDirFS = apis.MustSubFS(distDir, "dist")
