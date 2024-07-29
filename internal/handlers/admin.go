package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/forms"
)

func (app *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	app.view(w, r, "admin-dashboard.page.tmpl", nil)
}

func (app *Handlers) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := app.repo.GetNewReservations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.view(w, r, "admin-new-reservations.page.tmpl", map[string]any{
		"reservations": reservations,
	})
}

func (app *Handlers) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := app.repo.GetAllReservations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.view(w, r, "admin-all-reservations.page.tmpl", map[string]any{
		"reservations": reservations,
	})
}

func (app *Handlers) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reservation, err := app.repo.GetReservationById(resId)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			app.Session.Put(r.Context(), "error", "Reservation doesn't exist")
			http.Redirect(w, r, "/admin/reservations/all", http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.view(w, r, "admin-show-reservation.page.tmpl", map[string]any{
		"reservation": reservation,
		"csrf":        csrfToken(w),
	})
}

func (app *Handlers) AdminPostEditReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reservation, err := app.repo.GetReservationById(resId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newToken := csrfToken(w)

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		app.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		app.view(w, r, "admin-show-reservation.page.tmpl", map[string]any{
			"reservation": reservation,
			"errors":      form.Errors,
			"csrf":        newToken,
		})
		return
	}

	if err = app.repo.UpdateReservation(reservation); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Session.Put(r.Context(), "flash", "Record Updated!")
	http.Redirect(w, r, "/admin/reservations/all", http.StatusSeeOther)
}

func (app *Handlers) AdminPostApproveReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	processed, _ := strconv.Atoi(r.FormValue("is-approved"))
	if processed == 1 {
		if err = app.repo.ApproveReservation(resId, 0); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		app.Session.Put(r.Context(), "flash", "Reservation Unapproved!")
	} else {
		if err = app.repo.ApproveReservation(resId, 1); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		app.Session.Put(r.Context(), "flash", "Reservation Approved!")
	}

	http.Redirect(w, r, "/admin/reservations/new", http.StatusSeeOther)
}

func (app *Handlers) AdminPostDeleteReservation(w http.ResponseWriter, r *http.Request) {
	resId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = app.repo.DeleteReservation(resId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, "/admin/reservations/all", http.StatusSeeOther)
}
