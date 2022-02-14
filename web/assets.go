package web

import "embed"

var (
	//go:embed static
	//go:embed template
	StaticFS embed.FS
)
