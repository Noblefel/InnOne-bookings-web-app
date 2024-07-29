package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAdminNewReservations(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.AdminNewReservations(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestAdminAllReservations(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.AdminAllReservations(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestAdminShowReservations(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.AdminShowReservation(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestAdminPostEditReservation(t *testing.T) {
	form := url.Values{
		"first_name": {"abcdefg"},
		"last_name":  {"abcdefg"},
		"email":      {"abcdefg@example.com"},
		"phone":      {"12345"},
	}.Encode()

	req := httptest.NewRequest("POST", "/", strings.NewReader(form))
	req.SetPathValue("id", "1")
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.AdminPostEditReservation(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}
}

func TestAdminPostApproveReservation(t *testing.T) {
	form := url.Values{}.Encode()
	req := httptest.NewRequest("POST", "/", strings.NewReader(form))
	req.SetPathValue("id", "1")
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.AdminPostApproveReservation(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}
}

func TestAdminPostDeleteReservation(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.AdminPostDeleteReservation(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}
}
