package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
)

var reservation = types.Reservation{
	RoomId: 1,
	Room: types.Room{
		Id:       1,
		Slug:     "sovereigns-suite",
		RoomName: "Sovereign's Suite",
	},
	StartDate: time.Now(),
	EndDate:   time.Now().AddDate(0, 0, 1),
}

var room = types.Room{
	Id:   1,
	Slug: "sovereigns-suite",
}

func TestReservation(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(getCtx(req))
		rec := httptest.NewRecorder()
		handlers.Session.Put(req.Context(), "reservation", reservation)
		handlers.Reservation(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("want 200 ok, got %d", rec.Code)
		}
	})

	t.Run("redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(getCtx(req))
		rec := httptest.NewRecorder()
		handlers.Reservation(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Errorf("want temporary redirect, got %d", rec.Code)
		}
	})
}

func TestPostReservation(t *testing.T) {
	form := url.Values{
		"first_name": {"abcdefg"},
		"last_name":  {"abcdefg"},
		"email":      {"abcdefg@example.com"},
		"phone":      {"12345"},
	}.Encode()

	t.Run("ok", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(form))
		req = req.WithContext(getCtx(req))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handlers.Session.Put(req.Context(), "reservation", reservation)
		handlers.PostReservation(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("want status see other, got %d", rec.Code)
		}
	})

	t.Run("if room not available", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/", strings.NewReader(form))
		req = req.WithContext(getCtx(req))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		modified := reservation
		modified.StartDate = modified.EndDate.AddDate(1, 0, 0)
		handlers.Session.Put(req.Context(), "reservation", modified)
		handlers.PostReservation(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("want 409 conflict, got %d", rec.Code)
		}
	})
}

func TestReservationSummary(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.Session.Put(req.Context(), "reservation", reservation)
	handlers.ReservationSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestPostCheckAllAvailabilityJSON(t *testing.T) {
	form := url.Values{
		"start_date": {"2020-01-02"},
		"end_date":   {"2020-01-03"},
	}.Encode()

	req := httptest.NewRequest("POST", "/", strings.NewReader(form))
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.PostCheckAllAvailabilityJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestPostCheckAvailabilityJSON(t *testing.T) {
	form := url.Values{
		"start_date": {"2020-01-02"},
		"end_date":   {"2020-01-03"},
	}.Encode()

	req := httptest.NewRequest("POST", "/", strings.NewReader(form))
	req.SetPathValue("slug", room.Slug)
	req = req.WithContext(getCtx(req))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handlers.Session.Put(req.Context(), "room", room)
	handlers.PostCheckAvailabilityJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestRoom(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("slug", room.Slug)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.Room(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200 ok, got %d", rec.Code)
	}
}

func TestBookRoom(t *testing.T) {
	req := httptest.NewRequest("GET", "/?start_date=2020-01-01&end_date=2020-01-02", nil)
	req.SetPathValue("slug", room.Slug)
	req = req.WithContext(getCtx(req))
	rec := httptest.NewRecorder()
	handlers.BookRoom(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("want status see other, got %d", rec.Code)
	}
}
