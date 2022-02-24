package templates

import "embed"

var (
	//go:embed *.json
	EmbeddedTemplates embed.FS
)
