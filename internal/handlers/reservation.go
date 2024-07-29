package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/forms"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
)

func (app *Handlers) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := app.Session.Get(r.Context(), "reservation").(types.Reservation)
	if !ok {
		app.Session.Put(r.Context(), "error", "Check any availability first")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token := csrfToken(w)
	app.Session.Put(r.Context(), "reservation", res)
	app.view(w, r, "reservation.page.tmpl", map[string]any{
		"reservation": res,
		"csrf":        token,
	})
}

func (app *Handlers) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := app.Session.Get(r.Context(), "reservation").(types.Reservation)
	if !ok {
		app.Session.Put(r.Context(), "error", "Check any availability first")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	newToken := csrfToken(w)
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")
	app.Session.Put(r.Context(), "reservation", reservation)

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		app.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		app.view(w, r, "reservation.page.tmpl", map[string]any{
			"reservation": reservation,
			"errors":      form.Errors,
			"csrf":        newToken,
		})
		return
	}

	b, err := app.repo.SearchRoomAvailabilityByDates(reservation.StartDate, reservation.EndDate, reservation.RoomId)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Cannot check the room's availability")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !b {
		app.Session.Put(r.Context(), "error", "Room is not available")
		w.WriteHeader(http.StatusConflict)
		app.view(w, r, "reservation.page.tmpl", map[string]any{
			"reservation": reservation,
			"csrf":        newToken,
		})
		return
	}

	newReservationId, err := app.repo.InsertReservation(reservation)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Cannot insert reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := types.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
	}

	if err = app.repo.InsertRoomRestriction(restriction); err != nil {
		app.Session.Put(r.Context(), "error", "Cannot insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	msg := types.MailData{
		To:       reservation.Email,
		From:     "test@here.com",
		Subject:  "Reservation Confirmation",
		Template: "reservation-summary.tmpl",
		TemplateContent: map[string]string{
			"[%name%]": reservation.FirstName,
			"[%room%]": reservation.Room.RoomName,
			"[%from%]": reservation.StartDate.Format("2006-01-02"),
			"[%to%]":   reservation.EndDate.Format("2006-01-02"),
		},
	}

	app.MailChan <- msg
	http.Redirect(w, r, "reservation-summary", http.StatusSeeOther)
}

func (app *Handlers) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := app.Session.Get(r.Context(), "reservation").(types.Reservation)
	if !ok {
		app.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	app.Session.Remove(r.Context(), "reservation")
	app.view(w, r, "reservation-summary.page.tmpl", map[string]any{
		"reservation": reservation,
	})
}

func (app *Handlers) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	app.view(w, r, "check.page.tmpl", map[string]any{"csrf": csrfToken(w)})
}

func (app *Handlers) PostCheckAllAvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "")
		return
	}

	startDate, err := time.Parse("2006-01-02", r.Form.Get("start_date"))
	if err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing start date")
		return
	}

	endDate, err := time.Parse("2006-01-02", r.Form.Get("end_date"))
	if err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing end date")
		return
	}

	rooms, err := app.repo.SearchAllRoomsAvailibility(startDate, endDate)
	if err != nil {
		ServerErrorJSON(w, http.StatusInternalServerError, err, "Error searching room availability")
		return
	}

	res := JSONResponse{
		Ok:      true,
		Message: "Searched All Available Rooms",
		Data:    map[string][]types.Room{"rooms": rooms},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (app *Handlers) PostCheckAvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "")
		return
	}

	startDate, err := time.Parse("2006-01-02", r.Form.Get("start_date"))
	if err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing start date")
		return
	}

	endDate, err := time.Parse("2006-01-02", r.Form.Get("end_date"))
	if err != nil {
		ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing end date")
		return
	}

	room, ok := app.Session.Get(r.Context(), "room").(types.Room)
	if !ok {
		ServerErrorJSON(w, http.StatusInternalServerError, err, "Cannot get room from session, try refreshing the page")
		return
	}

	if room.Slug != r.PathValue("slug") {
		log.Println(room.Slug, r.PathValue("slug"))
		ServerErrorJSON(w, http.StatusInternalServerError, errors.New("session room slug doesn't match route"), "")
		return
	}

	b, err := app.repo.SearchRoomAvailabilityByDates(startDate, endDate, room.Id)
	if err != nil {
		ServerErrorJSON(w, http.StatusInternalServerError, err, "Error checking room availability")
		return
	}

	if !b {
		ServerErrorJSON(w, http.StatusNotFound, err, "Room is not available")
		return
	}

	res := JSONResponse{
		Ok:      true,
		Message: "Available",
		Data:    map[string]bool{"available": true},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (app *Handlers) Room(w http.ResponseWriter, r *http.Request) {
	room, err := app.repo.GetRoomBySlug(r.PathValue("slug"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error getting room data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if room == (types.Room{}) {
		app.Session.Put(r.Context(), "error", "Room doesn't exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "room", room)
	app.view(w, r, "room.page.tmpl", map[string]any{
		"room": room,
		"csrf": csrfToken(w),
	})
}

func (app *Handlers) BookRoom(w http.ResponseWriter, r *http.Request) {
	startDate, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error parsing start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error parsing end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := app.repo.GetRoomBySlug(r.PathValue("slug"))
	if err != nil {
		app.Session.Put(r.Context(), "error", "Error getting room data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if room == (types.Room{}) {
		app.Session.Put(r.Context(), "error", "Room doesn't exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "reservation", types.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		Room:      room,
		RoomId:    room.Id,
	})

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
