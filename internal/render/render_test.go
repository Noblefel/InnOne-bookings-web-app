package render

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
)

var pageTemplate = []byte(`{{template "base" .}}{{define "body"}}{{index .Page "key"}}{{end}}`)
var layoutTemplate = []byte(`{{define "base"}}{{block "body" .}}{{end}}---{{end}}`)

func TestCacheRenderer(t *testing.T) {
	fsys := fstest.MapFS{
		"a.page.tmpl":   {Data: pageTemplate},
		"b.layout.tmpl": {Data: layoutTemplate},
	}

	r, err := New(fsys)
	if err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	data := types.TemplateData{Page: map[string]any{"key": "abc"}}

	if err := r.View(&sb, "a.page.tmpl", &data); err != nil {
		t.Fatal(err)
	}

	want := "abc---"
	if sb.String() != want {
		t.Errorf("incorrect template result, want %q, got %q", want, sb.String())
	}
}

func TestNoCacheRenderer(t *testing.T) {
	dir := t.TempDir()
	page, err := os.Create(dir + "/a.page.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	page.Write(pageTemplate)
	page.Close()

	layout, err := os.Create(dir + "/b.layout.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	layout.Write(layoutTemplate)
	layout.Close()

	r := NewNoCache(dir)

	var sb strings.Builder
	data := types.TemplateData{Page: map[string]any{"key": "abc"}}

	if err := r.View(&sb, "a.page.tmpl", &data); err != nil {
		t.Fatal(err)
	}

	want := "abc---"
	if sb.String() != want {
		t.Errorf("incorrect template result, want %q, got %q", want, sb.String())
	}
}
