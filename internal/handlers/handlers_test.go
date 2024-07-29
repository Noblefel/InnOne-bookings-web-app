package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/repository"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	"github.com/alexedwards/scs/v2"
)

var handlers *Handlers

type mockRenderer struct{}

func (r *mockRenderer) View(io.Writer, string, *types.TemplateData) error {
	return nil
}

func init() {
	app := &types.App{}
	app.MailChan = make(chan types.MailData)
	app.Session = scs.New()
	repo := repository.NewMockRepo()
	renderer := &mockRenderer{}
	handlers = New(app, repo, renderer)

	go func() {
		<-app.MailChan
	}()
}

func getCtx(r *http.Request) context.Context {
	ctx, _ := handlers.Session.Load(r.Context(), "")
	return ctx
}

func TestBasicPages(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"home page", handlers.Home},
		{"about page", handlers.About},
		{"check availability page", handlers.CheckAvailability},
		{"login page", handlers.Login},
		{"admin dashboard", handlers.AdminDashboard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(getCtx(req))
			tt.handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("want 200 ok, got %d", rec.Code)
			}
		})
	}
}
func TestPostLogin(t *testing.T) {
	form := url.Values{
		"email":    {"abcdefg@example.com"},
		"password": {"12345"},
	}.Encode()

	req := httptest.NewRequest("POST", "/", strings.NewReader(form))
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.PostLogin(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}

	if handlers.Session.GetInt(req.Context(), "auth_id") == 0 {
		t.Error("expecting auth id")
	}
}

func TestLogout(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.Session.Put(req.Context(), "key", "abc")
	handlers.Logout(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}

	if handlers.Session.GetString(req.Context(), "key") != "" {
		t.Error("expecting session to be refreshed")
	}
}
