package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/handlers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	"github.com/alexedwards/scs/v2"
)

type mockRenderer struct{}

func (r *mockRenderer) View(w io.Writer, tmpl string, d *types.TemplateData) error {
	return nil
}

func TestRoutes(t *testing.T) {
	app := &types.App{Session: scs.New()}
	renderer := &mockRenderer{}
	handlers := handlers.New(app, nil, renderer)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/404", nil)
	routes(handlers).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("invalid route should return 404, got %d", rec.Code)
	}
}
