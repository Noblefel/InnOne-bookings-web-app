package render

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"text/template"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
)

var funcs = template.FuncMap{
	"humanDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}

type Renderer interface {
	View(w io.Writer, tmpl string, data *types.TemplateData) error
}

type CacheRenderer struct{ cache map[string]*template.Template }

func New(fsys fs.FS) (*CacheRenderer, error) {
	matches, err := fs.Glob(fsys, "*.page.tmpl")
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*template.Template)

	for _, m := range matches {
		t, err := template.New(m).Funcs(funcs).ParseFS(fsys, m, "*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		cache[m] = t
	}

	return &CacheRenderer{cache}, nil
}

func (r *CacheRenderer) View(w io.Writer, tmpl string, data *types.TemplateData) error {
	t, ok := r.cache[tmpl]
	if !ok {
		return errors.New("no template found")
	}

	return write(w, t, data)
}

type NoCacheRenderer struct{ dir string }

func NewNoCache(dir string) *NoCacheRenderer {
	return &NoCacheRenderer{dir}
}

func (r *NoCacheRenderer) View(w io.Writer, tmpl string, data *types.TemplateData) error {
	fsys := os.DirFS(r.dir)

	matches, err := fs.Glob(fsys, tmpl)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return errors.New("no template found")
	}

	t, err := template.New(tmpl).Funcs(funcs).ParseFS(fsys, matches[0], "*.layout.tmpl")
	if err != nil {
		return err
	}

	return write(w, t, data)
}

func write(w io.Writer, t *template.Template, data *types.TemplateData) error {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}
