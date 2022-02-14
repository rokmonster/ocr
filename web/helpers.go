package web

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Masterminds/sprig/v3"
)

func CreateTemplateEngine(embedFS embed.FS, subpath string) *template.Template {
	root := template.New("").Funcs(sprig.FuncMap())
	sub, _ := fs.Sub(embedFS, subpath)

	err := fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			f, errOpen := sub.Open(path)
			if errOpen != nil {
				return errOpen
			}
			b, errRead := io.ReadAll(f)
			if errRead != nil {
				return errRead
			}

			t := root.New(path)
			_, errParse := t.Parse(string(b))
			if errParse != nil {
				return errParse
			}
		}

		return nil
	})

	return template.Must(root, err)
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
