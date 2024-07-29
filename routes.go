package main

import (
	"io/fs"
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/handlers"
)

func routes(h *handlers.Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /check-availability", h.CheckAvailability)
	mux.HandleFunc("POST /check-availability-all", h.PostCheckAllAvailabilityJSON)
	mux.HandleFunc("POST /check-availability/{slug}", h.PostCheckAvailabilityJSON)
	mux.HandleFunc("GET /make-reservation", h.Reservation)
	mux.HandleFunc("POST /make-reservation", h.PostReservation)
	mux.HandleFunc("GET /reservation-summary", h.ReservationSummary)
	mux.HandleFunc("GET /rooms/{slug}", h.Room)
	mux.HandleFunc("GET /rooms/{slug}/book", h.BookRoom)

	mux.HandleFunc("GET /logout", h.Logout)
	mux.Handle("GET /login", Guest(h.Login))
	mux.Handle("POST /login", Guest(h.PostLogin))

	mux.Handle("GET /admin", http.RedirectHandler("/admin/dashboard", http.StatusSeeOther))
	mux.Handle("GET /admin/dashboard", Auth(h.AdminDashboard))
	mux.Handle("GET /admin/reservations/new", Auth(h.AdminNewReservations))
	mux.Handle("GET /admin/reservations/all", Auth(h.AdminAllReservations))
	mux.Handle("GET /admin/reservations/{id}", Auth(h.AdminShowReservation))
	mux.Handle("POST /admin/reservations/{id}/edit", Auth(h.AdminPostEditReservation))
	mux.Handle("POST /admin/reservations/{id}/approve", Auth(h.AdminPostApproveReservation))
	mux.Handle("POST /admin/reservations/{id}/delete", Auth(h.AdminPostDeleteReservation))

	mux.HandleFunc("/{$}", h.Home)
	mux.Handle("/", http.NotFoundHandler())

	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	staticFS, _ := fs.Sub(staticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	return app.Session.LoadAndSave(CSRF(mux))
}
