package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	"github.com/alexedwards/scs/v2"
)

func init() {
	app = &types.App{}
	app.Session = scs.New()
}

func TestCSRF(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", nil)
		req.AddCookie(&http.Cookie{Name: "csrf", Value: "abc"})
		req.Form = url.Values{"csrf": {"abc"}}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handler := CSRF(next)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("want %d status code, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("should unauthorized", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", nil)
		req.AddCookie(&http.Cookie{Name: "csrf", Value: "abc"})

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handler := CSRF(next)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("want %d status code, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestAuth(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx, _ := app.Session.Load(req.Context(), "")
	req = req.WithContext(ctx)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := Auth(next)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want %d status code, got %d", http.StatusSeeOther, rec.Code)
	}
}

func TestGuest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx, _ := app.Session.Load(req.Context(), "")
	req = req.WithContext(ctx)
	app.Session.Put(req.Context(), "auth_id", 1)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := Guest(next)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want %d status code, got %d", http.StatusSeeOther, rec.Code)
	}
}
