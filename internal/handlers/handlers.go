package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/driver"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/forms"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/helpers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/render"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/repository"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// The Repository used by the handlers
var Repo *Repository

// is the Repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(ac *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: ac,
		DB:  dbrepo.NewPostgresRepo(db.SQL, ac),
	}
}

func NewTestRepo(ac *config.AppConfig) *Repository {
	return &Repository{
		App: ac,
		DB:  dbrepo.NewTestingRepo(ac),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (x *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (x *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (x *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	res, ok := x.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		x.App.Session.Put(r.Context(), "error", "Check any availability first")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	startDate := res.StartDate.Format("2006-01-02")
	endDate := res.EndDate.Format("2006-01-02")
	x.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: map[string]interface{}{
			"reservation": res,
		},
		StringMap: map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

func (x *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := x.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		x.App.Session.Put(r.Context(), "error", "Check any availability first")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Can't pass form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	x.App.Session.Put(r.Context(), "reservation", reservation)

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		x.App.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		render.Template(w, r, "reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: map[string]interface{}{
				"reservation": reservation,
			},
			StringMap: map[string]string{
				"start_date": startDate,
				"end_date":   endDate,
			},
		})
		return
	}

	b, err := x.DB.SearchRoomAvailabilityByDates(reservation.StartDate, reservation.EndDate, reservation.RoomId)
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Cannot check the room's availability")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !b {
		x.App.Session.Put(r.Context(), "error", "Room is not available")
		w.WriteHeader(http.StatusConflict)
		render.Template(w, r, "reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: map[string]interface{}{
				"reservation": reservation,
			},
			StringMap: map[string]string{
				"start_date": startDate,
				"end_date":   endDate,
			},
		})
		return
	}

	newReservationId, err := x.DB.InsertReservation(reservation)
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Cannot insert reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomId:        reservation.RoomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
	}

	err = x.DB.InsertRoomRestriction(restriction)
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Cannot insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notification to the guest
	msg := models.MailData{
		To:       reservation.Email,
		From:     "test@here.com",
		Subject:  "Reservation Confirmation",
		Template: "reservation-summary.tmpl",
		TemplateContent: map[string]string{
			"[%name%]": reservation.FirstName,
			"[%room%]": reservation.Room.RoomName,
			"[%from%]": startDate,
			"[%to%]":   endDate,
		},
	}

	x.App.MailChan <- msg

	http.Redirect(w, r, "reservation-summary", http.StatusSeeOther)
}

func (x *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {

	reservation, ok := x.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		x.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	x.App.Session.Remove(r.Context(), "reservation")

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservation": reservation,
		},
		StringMap: map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

func (x *Repository) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "check.page.tmpl", &models.TemplateData{})
}

func (x *Repository) PostCheckAllAvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "")
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.Form.Get("start_date"))
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing start date")
		return
	}

	endDate, err := time.Parse(layout, r.Form.Get("end_date"))
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing end date")
		return
	}

	rooms, err := x.DB.SearchAllRoomsAvailibility(startDate, endDate)
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, err, "Error searching room availability")
		return
	}

	sort.SliceStable(rooms, func(i, j int) bool {
		return rooms[i].Slug < rooms[j].Slug
	})

	res := helpers.JSONResponse{
		Ok:      true,
		Message: "Searched All Available Rooms",
		Data:    map[string][]models.Room{"rooms": rooms},
	}

	json, err := json.Marshal(res)
	if err != nil || r.Form.Get("TEST_ERROR") != "" {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, err, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (x *Repository) PostCheckAvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "")
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.Form.Get("start_date"))
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing start date")
		return
	}

	endDate, err := time.Parse(layout, r.Form.Get("end_date"))
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusBadRequest, err, "Error parsing end date")
		return
	}

	room, ok := x.App.Session.Get(r.Context(), "room").(models.Room)
	if !ok {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, err, "Cannot get room from session, try refreshing the page")
		return
	}

	if room.Slug != chi.URLParam(r, "slug") {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, errors.New("Session room slug doesn't match route"), "")
		return
	}

	b, err := x.DB.SearchRoomAvailabilityByDates(startDate, endDate, room.Id)
	if err != nil {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, err, "Error checking room availability")
		return
	}

	if !b {
		helpers.ServerErrorJSON(w, http.StatusNotFound, err, "Room is not available")
		return
	}

	res := helpers.JSONResponse{
		Ok:      true,
		Message: "Available",
		Data:    map[string]bool{"available": true},
	}

	json, err := json.Marshal(res)
	if err != nil || r.Form.Get("TEST_ERROR") != "" {
		helpers.ServerErrorJSON(w, http.StatusInternalServerError, err, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (x *Repository) Room(w http.ResponseWriter, r *http.Request) {

	room, err := x.DB.GetRoomBySlug(chi.URLParam(r, "slug"))
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Error getting room data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if room == (models.Room{}) {
		x.App.Session.Put(r.Context(), "error", "Room doesn't exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	x.App.Session.Put(r.Context(), "room", room)

	render.Template(w, r, "room.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"room": room,
		},
	})
}

func (x *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.URL.Query().Get("start_date"))
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Error parsing start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, r.URL.Query().Get("end_date"))
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Error parsing end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := x.DB.GetRoomBySlug(chi.URLParam(r, "slug"))
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Error getting room data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if room == (models.Room{}) {
		x.App.Session.Put(r.Context(), "error", "Room doesn't exist")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	x.App.Session.Put(r.Context(), "reservation", models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		Room:      room,
		RoomId:    room.Id,
	})

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

func (x *Repository) Login(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (x *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		x.App.Session.Put(r.Context(), "error", "Can't pass form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		x.App.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := x.DB.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
	if err != nil {
		if errors.Is(bcrypt.ErrMismatchedHashAndPassword, err) {
			x.App.Session.Put(r.Context(), "error", "Invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			render.Template(w, r, "login.page.tmpl", &models.TemplateData{
				Form: form,
			})
			return
		}

		x.App.Session.Put(r.Context(), "error", "Something went wrong, cannot authenticate")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	x.App.Session.Put(r.Context(), "user_id", id)

	x.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (x *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = x.App.Session.Destroy(r.Context())
	_ = x.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (x *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (x *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := x.DB.GetNewReservations()
	if err != nil || strings.HasSuffix(r.URL.Path, "TEST_ERROR") {
		helpers.ServerError(w, err)
		return
	}

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservations": reservations,
		},
	})
}

func (x *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := x.DB.GetAllReservations()
	if err != nil || strings.HasSuffix(r.URL.Path, "TEST_ERROR") {
		helpers.ServerError(w, err)
		return
	}

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservations": reservations,
		},
	})
}

func (x *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{})
}

func (x *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation, err := x.DB.GetReservationById(resId)
	if err != nil {

		if errors.Is(sql.ErrNoRows, err) {
			x.App.Session.Put(r.Context(), "error", "Reservation doesn't exist")
			http.Redirect(w, r, "/admin/reservations/all", http.StatusTemporaryRedirect)
			return
		}

		helpers.ServerError(w, err)
		return
	}

	render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: map[string]interface{}{
			"reservation": reservation,
		},
	})
}

func (x *Repository) AdminPostEditReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation, err := x.DB.GetReservationById(resId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		x.App.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: map[string]interface{}{
				"reservation": reservation,
			},
		})
		return
	}

	err = x.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	x.App.Session.Put(r.Context(), "flash", "Record Updated!")
	http.Redirect(w, r, "/admin/reservations/all", http.StatusSeeOther)
}

func (x Repository) AdminPostApproveReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	processed, _ := strconv.Atoi(r.FormValue("is-approved"))
	if processed == 1 {
		err = x.DB.ApproveReservation(resId, 0)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		x.App.Session.Put(r.Context(), "flash", "Reservation Unapproved!")
	} else {
		err = x.DB.ApproveReservation(resId, 1)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		x.App.Session.Put(r.Context(), "flash", "Reservation Approved!")
	}

	http.Redirect(w, r, "/admin/reservations/new", http.StatusSeeOther)
}

func (x Repository) AdminPostDeleteReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = x.DB.DeleteReservation(resId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	x.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, "/admin/reservations/all", http.StatusSeeOther)
}
