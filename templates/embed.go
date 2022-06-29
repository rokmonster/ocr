package templates

import "embed"

var (
	//go:embed *.json
	FS embed.FS
)
