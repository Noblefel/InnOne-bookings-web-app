package main

import (
	"testing"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// Passed
	default:
		t.Errorf("Type is not *chi.Mux, but %T", v)
	}
}
