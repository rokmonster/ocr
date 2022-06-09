package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofiber/template/html"
)

func CreateTemplateEngine(embedFS embed.FS, subpath string) *html.Engine {
	sub, _ := fs.Sub(embedFS, subpath)
	engine := html.NewFileSystem(http.FS(sub), ".html")

	for name, f := range sprig.HtmlFuncMap() {
		engine.AddFunc(name, f)
	}

	return engine
}

type binaryFileSystem struct {
	fs http.FileSystem
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.fs.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

func EmbeddedFS(embedFS embed.FS, subdir string) *binaryFileSystem {
	sub, _ := fs.Sub(embedFS, subdir)
	return &binaryFileSystem{
		fs: http.FS(sub),
	}
}
